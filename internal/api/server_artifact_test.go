package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestArtifactEndpointPermissionsAndMissingIdentity(t *testing.T) {
	f := newReliabilityFixture(t)
	if got := serverActionPermission("GET", "artifact"); got != "files.view" {
		t.Fatal(got)
	}
	for _, user := range []string{"admin", "owner", "subuser"} {
		w := f.request("GET", "/api/servers/alpha/artifact", "", user)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"status":"missing"`) {
			t.Fatalf("%s: %d %s", user, w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("identity response cacheable")
		}
		if strings.Contains(w.Body.String(), f.manager.GetWorkingDir("alpha")) {
			t.Fatal("host path leaked")
		}
	}
	if w := f.request("GET", "/api/servers/alpha/artifact", "", "stranger"); w.Code != 403 && w.Code != 404 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestArtifactCreateDownloadFailureIsReportedAndRolledBack(t *testing.T) {
	f := newReliabilityFixture(t)
	before, _ := f.db.ListServers()
	original := http.DefaultTransport
	http.DefaultTransport = catalogTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("unavailable")), Header: make(http.Header), Request: r}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	w := f.request("POST", "/api/servers", `{"name":"failed-install","type":"paper","version":"1.21.4","port":30555,"auto_download":true}`, "admin")
	if w.Code != 502 || !strings.Contains(w.Body.String(), "no server was created") {
		t.Fatal(w.Code, w.Body.String())
	}
	after, _ := f.db.ListServers()
	if len(after) != len(before) {
		t.Fatal("failed synchronous install left a server record")
	}
}

func TestArtifactRefreshValidation(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, query := range []string{"?refresh=true", "?refresh=false"} {
		w := f.request("GET", "/api/servers/alpha/artifact"+query, "", "owner")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := f.request("GET", "/api/servers/alpha/artifact?refresh=arbitrary", "", "owner"); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
}
