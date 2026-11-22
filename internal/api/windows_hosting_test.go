package api

import (
	"encoding/json"
	"runtime"
	"testing"
)

func TestCreateServerUsesJava25Default(t *testing.T) {
	f := newReliabilityFixture(t)
	f.cfg.Java.DefaultPath = `C:\Program Files\Java\bin\java.exe`
	for i, tc := range []struct{ name, engine, javaPath, want string }{
		{"default", "custom", "", "25"},
		{"explicit", "custom", "java-override", "java-override"},
		{"legacy-runtime", "paper", "17", "17"},
		{"native", "steel", "java-ignored", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{"name": tc.name, "type": tc.engine, "java_path": tc.javaPath, "auto_download": false, "port": 30000 + i})
			w := f.request("POST", "/api/servers", string(body), "admin")
			if w.Code != 201 {
				t.Fatalf("%d: %s", w.Code, w.Body.String())
			}
			var result ServerWithStats
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.JavaPath != tc.want {
				t.Fatalf("Java path = %q, want %q", result.JavaPath, tc.want)
			}
			if result.Stats.NetworkSupported == nil || *result.Stats.NetworkSupported != (runtime.GOOS == "linux") {
				t.Fatalf("missing or incorrect network capability: %+v", result.Stats)
			}
			saved, err := f.db.GetServer(result.ID)
			if err != nil || saved.JavaPath != tc.want {
				t.Fatalf("persisted Java path: %+v, %v", saved, err)
			}
		})
	}
}
