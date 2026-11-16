package webapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
)

func newAuthorizedClient(t *testing.T, baseURL, token string) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	response, err := client.Get(baseURL + "/?t=" + token)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("index status = %d, want 200", response.StatusCode)
	}
	return client
}

func TestIndexRequiresSession(t *testing.T) {
	app, err := New("secret")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(app)
	defer server.Close()

	response, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.StatusCode)
	}

	apiResponse, err := http.Get(server.URL + "/api/live")
	if err != nil {
		t.Fatal(err)
	}
	defer apiResponse.Body.Close()
	if apiResponse.StatusCode != http.StatusForbidden {
		t.Fatalf("api status = %d, want 403", apiResponse.StatusCode)
	}
}

func TestIndexTokenFlowAndAssets(t *testing.T) {
	app, err := New("secret")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(app)
	defer server.Close()

	client := newAuthorizedClient(t, server.URL, "secret")
	response, err := client.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if csp := response.Header.Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") {
		t.Fatalf("missing CSP, got %q", csp)
	}
	body, _ := io.ReadAll(response.Body)
	if !bytes.Contains(body, []byte("/static/styles.css")) ||
		!bytes.Contains(body, []byte("id=\"dropzone\"")) {
		t.Fatal("index body does not look like the application shell")
	}

	for path, wantType := range map[string]string{
		"/static/styles.css":  "text/css",
		"/static/app.js":      "text/javascript",
		"/static/favicon.svg": "image/svg+xml",
	} {
		assetResponse, err := client.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		if assetResponse.StatusCode != http.StatusOK {
			t.Fatalf("%s status = %d", path, assetResponse.StatusCode)
		}
		if contentType := assetResponse.Header.Get("Content-Type"); !strings.Contains(contentType, wantType) {
			t.Fatalf("%s content type = %q, want %q", path, contentType, wantType)
		}
		assetResponse.Body.Close()
	}

	missing, err := client.Get(server.URL + "/static/nope.css")
	if err != nil {
		t.Fatal(err)
	}
	missing.Body.Close()
	if missing.StatusCode != http.StatusNotFound {
		t.Fatalf("missing asset status = %d, want 404", missing.StatusCode)
	}
}

func TestHashAPI(t *testing.T) {
	app, err := New("secret")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(app)
	defer server.Close()
	client := newAuthorizedClient(t, server.URL, "secret")

	data := []byte("prismatic tactility")
	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/hash", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("X-Algorithms", "MD5,sha256,SM3,CRC32")
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d: %s", response.StatusCode, body)
	}

	var payload hashResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Size != int64(len(data)) {
		t.Fatalf("size = %d, want %d", payload.Size, len(data))
	}
	wantSHA := sha256.Sum256(data)
	if payload.Hashes["SHA256"] != hex.EncodeToString(wantSHA[:]) {
		t.Fatalf("SHA256 = %s", payload.Hashes["SHA256"])
	}
	if len(payload.Hashes) != 4 {
		t.Fatalf("got %d hashes, want 4", len(payload.Hashes))
	}

	bad, err := http.NewRequest(http.MethodPost, server.URL+"/api/hash", strings.NewReader("x"))
	if err != nil {
		t.Fatal(err)
	}
	bad.Header.Set("X-Algorithms", "NOPE")
	badResponse, err := client.Do(bad)
	if err != nil {
		t.Fatal(err)
	}
	defer badResponse.Body.Close()
	if badResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad algorithm status = %d, want 400", badResponse.StatusCode)
	}

	empty, err := http.NewRequest(http.MethodPost, server.URL+"/api/hash", strings.NewReader("x"))
	if err != nil {
		t.Fatal(err)
	}
	emptyResponse, err := client.Do(empty)
	if err != nil {
		t.Fatal(err)
	}
	defer emptyResponse.Body.Close()
	if emptyResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty algorithm status = %d, want 400", emptyResponse.StatusCode)
	}

	getResponse, err := client.Get(server.URL + "/api/hash")
	if err != nil {
		t.Fatal(err)
	}
	defer getResponse.Body.Close()
	if getResponse.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want 405", getResponse.StatusCode)
	}
}
