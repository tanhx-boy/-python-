// Package webapp serves the embedded user interface and hashing API on loopback.
package webapp

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"

	"filechecker/internal/hashkit"
)

//go:embed web/index.html web/static/*
var embedded embed.FS

const cookieName = "filechecker_session"

// exitGrace is how long the process lingers after the last browser window
// disconnects. It gives a page reload time to reconnect before shutdown.
const exitGrace = 8 * time.Second

type asset struct {
	data        []byte
	contentType string
}

// App implements the loopback UI server.
type App struct {
	token  string
	index  []byte
	assets map[string]asset

	liveMu    sync.Mutex
	liveCount int
	liveSeen  bool
	liveTimer *time.Timer

	exitOnce sync.Once
	exitCh   chan struct{}
}

// New loads the embedded UI and prepares the app. token must be presented as
// a query parameter once, then the browser keeps it as a session cookie.
func New(token string) (*App, error) {
	index, err := fs.ReadFile(embedded, "web/index.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded index: %w", err)
	}

	assetFiles := map[string]string{
		"/static/styles.css":  "text/css; charset=utf-8",
		"/static/app.js":      "text/javascript; charset=utf-8",
		"/static/favicon.svg": "image/svg+xml",
	}
	assets := make(map[string]asset, len(assetFiles))
	for path, contentType := range assetFiles {
		data, err := fs.ReadFile(embedded, "web"+path)
		if err != nil {
			return nil, fmt.Errorf("read embedded asset %s: %w", path, err)
		}
		assets[path] = asset{data: data, contentType: contentType}
	}

	return &App{
		token:  token,
		index:  index,
		assets: assets,
		exitCh: make(chan struct{}),
	}, nil
}

// ExitCh is closed when the application should terminate.
func (a *App) ExitCh() <-chan struct{} { return a.exitCh }

// Exit requests application shutdown. It is safe to call multiple times.
func (a *App) Exit() {
	a.exitOnce.Do(func() { close(a.exitCh) })
}

// HasEverConnected reports whether a UI window has connected at least once.
func (a *App) HasEverConnected() bool {
	a.liveMu.Lock()
	defer a.liveMu.Unlock()
	return a.liveSeen
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("X-Frame-Options", "DENY")

	switch {
	case r.URL.Path == "/":
		a.handleIndex(w, r)
	case strings.HasPrefix(r.URL.Path, "/static/"):
		a.handleAsset(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/"):
		if !a.authorized(r) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		a.handleAPI(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *App) authorized(r *http.Request) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil || len(cookie.Value) != len(a.token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(a.token)) == 1
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if token := r.URL.Query().Get("t"); token != "" &&
		subtle.ConstantTimeCompare([]byte(token), []byte(a.token)) == 1 {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    a.token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !a.authorized(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, forbiddenPage)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Security-Policy",
		"default-src 'self'; img-src 'self' data:; style-src 'self'; "+
			"script-src 'self'; connect-src 'self'; font-src 'self'; "+
			"base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	_, _ = w.Write(a.index)
}

func (a *App) handleAsset(w http.ResponseWriter, r *http.Request) {
	item, ok := a.assets[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	h := w.Header()
	h.Set("Content-Type", item.contentType)
	h.Set("Cache-Control", "no-store")
	h.Set("Cross-Origin-Resource-Policy", "same-origin")
	_, _ = w.Write(item.data)
}

func (a *App) handleAPI(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/hash":
		a.handleHash(w, r)
	case "/api/live":
		a.handleLive(w, r)
	default:
		http.NotFound(w, r)
	}
}

type hashResponse struct {
	Size   int64             `json:"size"`
	Hashes map[string]string `json:"hashes"`
}

func (a *App) handleHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	names := strings.Split(r.Header.Get("X-Algorithms"), ",")
	if _, err := hashkit.Resolve(names); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashes, size, err := hashkit.Hash(r.Body, names)
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read the uploaded file: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, hashResponse{Size: size, Hashes: hashes})
}

func (a *App) handleLive(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming is not supported")
		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-store")
	h.Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "retry: 2000\n\n")
	flusher.Flush()

	a.liveStarted()
	defer a.liveEnded()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if _, err := io.WriteString(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (a *App) liveStarted() {
	a.liveMu.Lock()
	defer a.liveMu.Unlock()
	a.liveCount++
	a.liveSeen = true
	if a.liveTimer != nil {
		a.liveTimer.Stop()
		a.liveTimer = nil
	}
}

func (a *App) liveEnded() {
	a.liveMu.Lock()
	defer a.liveMu.Unlock()
	a.liveCount--
	if a.liveCount > 0 {
		return
	}
	a.liveCount = 0
	if a.liveTimer != nil {
		a.liveTimer.Stop()
	}
	a.liveTimer = time.AfterFunc(exitGrace, a.Exit)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

const forbiddenPage = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>FileChecker</title></head>
<body style="font-family:Segoe UI,Microsoft YaHei,sans-serif;padding:48px;color:#22383b;background:#e5edf0">
<h1 style="font-size:18px">请从应用窗口打开 / Open from the app window</h1>
<p style="font-size:13px;color:#476267">这个页面受本地会话保护。请重新运行 FileChecker，或使用带访问令牌的链接。</p>
<p style="font-size:13px;color:#476267">This page is protected by a local session. Relaunch FileChecker to get a new link.</p>
</body></html>`
