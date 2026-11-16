// Command filechecker is a local file integrity checker with a prismatic,
// glass-tinted web UI served over loopback.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"filechecker/internal/host"
	"filechecker/internal/webapp"
)

const version = "2.0.0"

func main() {
	port := flag.Int("port", 0, "loopback TCP port (0 picks a free port)")
	noOpen := flag.Bool("no-open", false, "do not launch a browser window")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("FileChecker", version)
		return
	}

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		fail(fmt.Errorf("generate session token: %w", err))
	}
	token := hex.EncodeToString(tokenBytes)

	app, err := webapp.New(token)
	if err != nil {
		fail(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fail(fmt.Errorf("listen on loopback: %w", err))
	}
	url := fmt.Sprintf("http://%s/?t=%s", listener.Addr(), token)

	server := &http.Server{
		Handler:           app,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil &&
			!errors.Is(serveErr, http.ErrServerClosed) {
			log.Printf("http server: %v", serveErr)
			app.Exit()
		}
	}()

	fmt.Printf("FileChecker %s\nURL: %s\n", version, url)

	if !*noOpen {
		if err := host.Open(url); err != nil {
			log.Printf("could not open a browser window: %v", err)
			host.Notice("FileChecker",
				"无法自动打开浏览器窗口 / Could not open a browser window.\n\n请手动访问 / Open manually:\n"+url)
		}
		// If no UI ever connects (for example the browser was blocked), do not
		// linger forever as an invisible background process.
		time.AfterFunc(10*time.Minute, func() {
			if !app.HasEverConnected() {
				app.Exit()
			}
		})
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signals:
	case <-app.ExitCh():
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func fail(err error) {
	log.Printf("fatal: %v", err)
	host.Notice("FileChecker", "启动失败 / Failed to start:\n\n"+err.Error())
	os.Exit(1)
}
