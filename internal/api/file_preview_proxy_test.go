package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestFilePreviewProxyDoesNotTruncateSlowMediaBody(t *testing.T) {
	f := newReliabilityFixture(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "4")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("a"))
		w.(http.Flusher).Flush()
		time.Sleep(120 * time.Millisecond)
		_, _ = w.Write([]byte("bcd"))
	}))
	defer upstream.Close()
	if err := f.db.CreateNode(&database.ClusterNode{ID: "media", Name: "Media", BaseURL: upstream.URL, APIToken: "node-token"}); err != nil {
		t.Fatal(err)
	}
	f.router.nodes.httpClient.Timeout = 30 * time.Millisecond
	w := f.request("GET", "/api/nodes/media/proxy/api/servers/alpha/files/media?path=clip.mp4", "", "admin")
	if w.Code != http.StatusOK || w.Body.String() != "abcd" {
		t.Fatalf("media truncated: %d %q", w.Code, w.Body.String())
	}
	if f.router.nodes.httpClient.Timeout != 30*time.Millisecond {
		t.Fatal("changed the timeout for non-media requests")
	}
}

func TestFilePreviewProxyTimeoutScope(t *testing.T) {
	base := &http.Client{Timeout: 55 * time.Second}
	for _, path := range []string{"api/servers/a/files/content", "api/servers/a/files/zip", "api/servers/a/files/media/other", "api/system/media"} {
		if fileMediaProxyClient(base, http.MethodGet, path) != base {
			t.Fatal("removed ordinary API timeout", path)
		}
	}
	if fileMediaProxyClient(base, http.MethodPost, "api/servers/a/files/media") != base {
		t.Fatal("removed mutation timeout")
	}
	stream := fileMediaProxyClient(base, http.MethodGet, "api/servers/a/files/media")
	if stream.Timeout != 0 || stream.Transport == nil {
		t.Fatal("streaming client not configured")
	}
	if fileMediaProxyTransport.ResponseHeaderTimeout <= 0 || fileMediaProxyTransport.TLSHandshakeTimeout <= 0 {
		t.Fatal("unbounded setup timeout")
	}
}
