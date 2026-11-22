package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateServerLatestDefaultsAndExplicitChoices(t *testing.T) {
	f := newReliabilityFixture(t)
	f.cfg.Java.DefaultPath = "17"
	old := http.DefaultTransport
	http.DefaultTransport = steelAPITransport(func(r *http.Request) (*http.Response, error) {
		t.Errorf("creation without download performed HTTP: %s", r.URL)
		return nil, fmt.Errorf("network disabled")
	})
	t.Cleanup(func() { http.DefaultTransport = old })
	port := 30000
	for _, kind := range []string{"", "paper", "purpur", "leaf", "leafmc", "folia", "fabric", "forge", "neoforge", "vanilla", "velocity", "waterfall", "bungeecord", "steel", "custom"} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/explicit=%v", kind, explicit), func(t *testing.T) {
				version, javaPath := " \t", " \t"
				wantVersion, wantJava := "latest", "25"
				if explicit {
					version, javaPath = "1.16.5", `C:\custom-java\bin\java.exe`
					wantVersion, wantJava = version, javaPath
				}
				if kind == "steel" {
					wantJava = ""
				}
				port++ // Each creation reserves a distinct node-local primary port.
				body, _ := json.Marshal(map[string]any{"name": "defaults test", "type": kind, "version": version, "java_path": javaPath, "auto_download": false, "port": port})
				w := f.request("POST", "/api/servers", string(body), "admin")
				if w.Code != 201 {
					t.Fatalf("%d: %s", w.Code, w.Body.String())
				}
				var result ServerWithStats
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				saved, err := f.db.GetServer(result.ID)
				if err != nil {
					t.Fatal(err)
				}
				if saved.Version != wantVersion || saved.JavaPath != wantJava {
					t.Fatalf("version/java = %q/%q; want %q/%q", saved.Version, saved.JavaPath, wantVersion, wantJava)
				}
			})
		}
	}
	// Creation must not migrate an existing server or reset manual runtime choices.
	existing, err := f.db.GetServer("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if existing.Version != "1.21" || existing.JavaPath != "java-custom" {
		t.Fatalf("existing server changed: %+v", existing)
	}
	w := f.request("PUT", "/api/servers/alpha", `{"name":"renamed"}`, "admin")
	if w.Code != 200 {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	existing, err = f.db.GetServer("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if existing.Version != "1.21" || existing.JavaPath != "java-custom" {
		t.Fatalf("update changed choices: %+v", existing)
	}
}
