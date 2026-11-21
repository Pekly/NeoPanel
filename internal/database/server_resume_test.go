package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func openServerResumeDB(t *testing.T, path string) *DB {
	t.Helper()
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func serverResumeRaw(t *testing.T, db *DB, baseDir string) string {
	t.Helper()
	_, key, err := serverResumeScope(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := db.conn.QueryRow(`SELECT value FROM panel_settings WHERE key = ?`, key).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestServerResumeRoundTripAndSingleConsumption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")
	db := openServerResumeDB(t, path)
	root := filepath.Join(t.TempDir(), "servers")
	input := []string{"zeta", "alpha", "zeta", "server-1_2.3"}
	original := append([]string(nil), input...)
	if err := db.SaveServerResumeState(root, input); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatal("save mutated caller IDs")
	}
	var state serverResumeState
	if err := json.Unmarshal([]byte(serverResumeRaw(t, db, root)), &state); err != nil {
		t.Fatal(err)
	}
	want := []string{"alpha", "server-1_2.3", "zeta"}
	scope, key, err := serverResumeScope(root)
	if err != nil || state.Version != 1 || state.Scope != scope || !reflect.DeepEqual(state.IDs, want) {
		t.Fatalf("stored state = %+v, scope error = %v", state, err)
	}
	if !strings.HasPrefix(key, "server_resume:") || len(key) != len("server_resume:")+64 {
		t.Fatalf("key is not prefixed SHA-256: %q", key)
	}
	// Age is deliberately irrelevant, including across database reopen/startup.
	if _, err := db.conn.Exec(`UPDATE panel_settings SET updated_at = '2000-01-01' WHERE key = ?`, key); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openServerResumeDB(t, path)
	got, err := db.TakeServerResumeState(root)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("take = %v, %v", got, err)
	}
	for n := 0; n < 3; n++ {
		if got, err := db.TakeServerResumeState(root); err != nil || got != nil {
			t.Fatalf("repeated take = %v, %v", got, err)
		}
	}
}

func TestServerResumeMissingAndOverwriteEmpty(t *testing.T) {
	db := openServerResumeDB(t, filepath.Join(t.TempDir(), "panel.db"))
	root := t.TempDir()
	if got, err := db.TakeServerResumeState(root); err != nil || got != nil {
		t.Fatalf("missing take = %v, %v", got, err)
	}
	for _, empty := range [][]string{nil, {}} {
		if err := db.SaveServerResumeState(root, []string{"old"}); err != nil {
			t.Fatal(err)
		}
		if err := db.SaveServerResumeState(root, empty); err != nil {
			t.Fatal(err)
		}
		if raw := serverResumeRaw(t, db, root); !strings.Contains(raw, `"ids":[]`) {
			t.Fatalf("empty snapshot not stored: %s", raw)
		}
		if got, err := db.TakeServerResumeState(root); err != nil || got == nil || len(got) != 0 {
			t.Fatalf("empty take = %#v, %v", got, err)
		}
		if got, err := db.TakeServerResumeState(root); err != nil || got != nil {
			t.Fatalf("empty snapshot not consumed: %v, %v", got, err)
		}
	}
	if err := db.SaveServerResumeState(root, []string{"old"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveServerResumeState(root, []string{"new"}); err != nil {
		t.Fatal(err)
	}
	if got, err := db.TakeServerResumeState(root); err != nil || !reflect.DeepEqual(got, []string{"new"}) {
		t.Fatalf("overwrite take = %v, %v", got, err)
	}
}

func TestServerResumeDirectoryScopes(t *testing.T) {
	db := openServerResumeDB(t, filepath.Join(t.TempDir(), "panel.db"))
	root := t.TempDir()
	a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
	if err := db.SaveServerResumeState(a, []string{"alpha"}); err != nil {
		t.Fatal(err)
	}
	if got, err := db.TakeServerResumeState(b); err != nil || got != nil {
		t.Fatalf("different root returned state: %v, %v", got, err)
	}
	if err := db.SaveServerResumeState(b, []string{"beta"}); err != nil {
		t.Fatal(err)
	}
	alias := a + string(filepath.Separator) + ".." + string(filepath.Separator) + "a" + string(filepath.Separator) + "."
	if runtime.GOOS == "windows" {
		alias = strings.ToUpper(alias)
	}
	if got, err := db.TakeServerResumeState(alias); err != nil || !reflect.DeepEqual(got, []string{"alpha"}) {
		t.Fatalf("canonical alias take = %v, %v", got, err)
	}
	if got, err := db.TakeServerResumeState(b); err != nil || !reflect.DeepEqual(got, []string{"beta"}) {
		t.Fatalf("other root take = %v, %v", got, err)
	}
	relScope, relKey, err := serverResumeScope(filepath.Join("resume-test", "..", "servers"))
	if err != nil || !filepath.IsAbs(relScope) {
		t.Fatalf("relative scope = %q, %v", relScope, err)
	}
	_, absKey, err := serverResumeScope(relScope)
	if err != nil || absKey != relKey {
		t.Fatalf("relative and absolute keys differ: %q, %q, %v", relKey, absKey, err)
	}
}

func TestServerResumeMalformedPreserved(t *testing.T) {
	db := openServerResumeDB(t, filepath.Join(t.TempDir(), "panel.db"))
	root := t.TempDir()
	scope, key, err := serverResumeScope(root)
	if err != nil {
		t.Fatal(err)
	}
	encodedScope, _ := json.Marshal(scope)
	valid := fmt.Sprintf(`{"version":1,"scope":%s,"ids":["alpha"]}`, encodedScope)
	cases := map[string]string{
		"syntax": "{", "null": "null", "array": "[]", "invalid-utf8": string([]byte{0xff}),
		"version":         strings.Replace(valid, `"version":1`, `"version":2`, 1),
		"missing-version": strings.Replace(valid, `"version":1,`, "", 1),
		"scope":           strings.Replace(valid, string(encodedScope), `"another-root"`, 1),
		"missing-scope":   `{"version":1,"ids":[]}`,
		"missing-ids":     fmt.Sprintf(`{"version":1,"scope":%s}`, encodedScope),
		"null-ids":        strings.Replace(valid, `["alpha"]`, `null`, 1),
		"null-id":         strings.Replace(valid, `["alpha"]`, `[null]`, 1),
		"bad-id":          strings.Replace(valid, `alpha`, `../escape`, 1),
		"control-id":      strings.Replace(valid, `alpha`, `bad\u0000id`, 1),
		"id-type":         strings.Replace(valid, `["alpha"]`, `[123]`, 1),
		"unknown-field":   strings.Replace(valid, `"version":1`, `"version":1,"extra":true`, 1),
		"trailing-json":   valid + `{}`, "trailing-junk": valid + `oops`,
	}
	tooMany := make([]string, maxServerResumeIDs+1)
	for i := range tooMany {
		tooMany[i] = "alpha"
	}
	payload, _ := json.Marshal(serverResumeState{Version: 1, Scope: scope, IDs: tooMany})
	cases["too-many"] = string(payload)
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := db.conn.Exec(`INSERT INTO panel_settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
				ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, raw); err != nil {
				t.Fatal(err)
			}
			for attempt := 0; attempt < 2; attempt++ {
				if got, err := db.TakeServerResumeState(root); err == nil || got != nil {
					t.Fatalf("corrupt take = %v, %v", got, err)
				}
				if got := serverResumeRaw(t, db, root); got != raw {
					t.Fatal("corrupt record changed")
				}
			}
		})
	}
}

func TestServerResumeRejectInvalidIDs(t *testing.T) {
	db := openServerResumeDB(t, filepath.Join(t.TempDir(), "panel.db"))
	root := t.TempDir()
	if err := db.SaveServerResumeState(root, []string{"original"}); err != nil {
		t.Fatal(err)
	}
	before := serverResumeRaw(t, db, root)
	invalid := []string{"", ".", "..", "a/b", `a\b`, "a:b", "a\x00b", "a\nb", "a\tb", "a\x7fb", "a\u0085b", strings.Repeat("x", 129), strings.Repeat("é", 65), string([]byte{0xff})}
	for _, id := range invalid {
		t.Run(fmt.Sprintf("%q", id), func(t *testing.T) {
			if err := db.SaveServerResumeState(root, []string{"valid", id}); err == nil {
				t.Fatalf("accepted invalid ID %q", id)
			}
			if serverResumeRaw(t, db, root) != before {
				t.Fatal("invalid save changed previous state")
			}
		})
	}
	ids := make([]string, maxServerResumeIDs+1)
	for i := range ids {
		ids[i] = fmt.Sprintf("server-%05d", i)
	}
	if err := db.SaveServerResumeState(root, ids); err == nil {
		t.Fatal("accepted too many IDs")
	}
	if serverResumeRaw(t, db, root) != before {
		t.Fatal("oversized save changed previous state")
	}
	ids = ids[:maxServerResumeIDs]
	ids[0] = strings.Repeat("x", 128)
	if err := db.SaveServerResumeState(root, ids); err != nil {
		t.Fatal("valid boundary rejected:", err)
	}
	if got, err := db.TakeServerResumeState(root); err != nil || len(got) != maxServerResumeIDs {
		t.Fatalf("boundary take count = %d, %v", len(got), err)
	}
}

func TestServerResumeTwoHandlesSingleConsumption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")
	first, second := openServerResumeDB(t, path), openServerResumeDB(t, path)
	root := t.TempDir()
	type result struct {
		ids []string
		err error
	}
	for attempt := 0; attempt < 30; attempt++ {
		want := []string{fmt.Sprintf("server-%d", attempt)}
		if err := first.SaveServerResumeState(root, want); err != nil {
			t.Fatal(err)
		}
		start, results := make(chan struct{}), make(chan result, 2)
		for _, db := range []*DB{first, second} {
			go func(db *DB) {
				<-start
				ids, err := db.TakeServerResumeState(root)
				results <- result{ids, err}
			}(db)
		}
		close(start)
		consumed := 0
		for n := 0; n < 2; n++ {
			got := <-results
			if got.err != nil && got.ids != nil {
				t.Fatalf("failed take exposed candidates: %+v", got)
			}
			if len(got.ids) != 0 {
				if !reflect.DeepEqual(got.ids, want) {
					t.Fatalf("unexpected IDs: %v", got.ids)
				}
				consumed++
			}
		}
		if consumed != 1 {
			t.Fatalf("snapshot consumed %d times", consumed)
		}
		for _, db := range []*DB{first, second} {
			if got, err := db.TakeServerResumeState(root); err != nil || got != nil {
				t.Fatalf("already consumed take = %v, %v", got, err)
			}
		}
	}
}

func TestServerResumeStorageFailuresPreserveState(t *testing.T) {
	db := openServerResumeDB(t, filepath.Join(t.TempDir(), "panel.db"))
	root := t.TempDir()
	if err := db.SaveServerResumeState(root, []string{"original"}); err != nil {
		t.Fatal(err)
	}
	before := serverResumeRaw(t, db, root)
	if _, err := db.conn.Exec(`CREATE TRIGGER resume_reject_update BEFORE UPDATE ON panel_settings
		BEGIN SELECT RAISE(ABORT, 'update blocked'); END;
		CREATE TRIGGER resume_reject_delete BEFORE DELETE ON panel_settings
		BEGIN SELECT RAISE(ABORT, 'delete blocked'); END;`); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveServerResumeState(root, nil); err == nil {
		t.Fatal("save hid storage failure")
	}
	if got, err := db.TakeServerResumeState(root); err == nil || got != nil {
		t.Fatalf("failed delete returned state: %v, %v", got, err)
	}
	if serverResumeRaw(t, db, root) != before {
		t.Fatal("storage failure changed snapshot")
	}
	if _, err := db.conn.Exec(`DROP TRIGGER resume_reject_update; DROP TRIGGER resume_reject_delete;`); err != nil {
		t.Fatal(err)
	}
	if got, err := db.TakeServerResumeState(root); err != nil || !reflect.DeepEqual(got, []string{"original"}) {
		t.Fatalf("take after failed transaction = %v, %v", got, err)
	}
	_, key, _ := serverResumeScope(root)
	var value string
	if err := db.conn.QueryRow(`SELECT value FROM panel_settings WHERE key = ?`, key).Scan(&value); err != sql.ErrNoRows {
		t.Fatalf("consumed record remains: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveServerResumeState(root, nil); err == nil {
		t.Fatal("closed DB save succeeded")
	}
	if got, err := db.TakeServerResumeState(root); err == nil || got != nil {
		t.Fatalf("closed DB take = %v, %v", got, err)
	}
}
