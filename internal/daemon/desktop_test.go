package daemon

import (
	"context"
	"net/http"
	"testing"
	"testing/fstest"
)

func TestDesktopFramingIsExplicitAndDocumentOnly(t *testing.T) {
	for _, desktop := range []bool{false, true} {
		cfg := testConfig(t)
		assets := fstest.MapFS{"index.html": {Data: []byte("panel")}}
		start := Start
		if desktop {
			start = StartDesktop
		}
		svc, err := start(cfg, assets)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{"/", "/index.html", "/api/auth/status", "/api/servers/example/files/preview"} {
			response, err := http.Get(svc.URL() + path)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			allowed := desktop && (path == "/" || path == "/index.html")
			if allowed {
				if response.Header.Get("X-Frame-Options") != "" || response.Header.Get("Content-Security-Policy") != "base-uri 'self'; frame-ancestors http://wails.localhost" {
					t.Fatal(path, response.Header)
				}
			} else if response.Header.Get("X-Frame-Options") != "DENY" {
				t.Fatal(path, response.Header)
			}
		}
		if err := svc.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDesktopRejectsExposedOrWorkerConfiguration(t *testing.T) {
	if _, err := StartDesktop(nil, nil); err == nil {
		t.Fatal("accepted nil config")
	}
	for _, kind := range []string{"network", "node", "sftp"} {
		cfg := testConfig(t)
		switch kind {
		case "network":
			cfg.Server.Host = "0.0.0.0"
		case "node":
			cfg.Server.Role = "node"
		case "sftp":
			cfg.SFTP.Enabled = true
		}
		if _, err := StartDesktop(cfg, nil); err == nil {
			t.Fatal("accepted", kind)
		}
	}
}
