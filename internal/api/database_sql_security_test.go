package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newSQLSecurityFixture(t *testing.T) (*reliabilityFixture, string) {
	t.Helper()
	f := newReliabilityFixture(t)
	if _, err := f.db.AddServerSubuser("alpha", f.users["stranger"].ID, []string{"database.manage"}); err != nil {
		t.Fatal(err)
	}
	w := f.request("POST", "/api/servers/alpha/databases", `{"name":"security_test"}`, "stranger")
	if w.Code != http.StatusCreated {
		t.Fatalf("create database: %d %s", w.Code, w.Body.String())
	}
	var record ServerDatabaseView
	if err := json.Unmarshal(w.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	return f, "/api/servers/alpha/databases/" + record.ID
}

func requestSecuritySQL(t *testing.T, f *reliabilityFixture, path, query string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(map[string]string{"sql": query})
	if err != nil {
		t.Fatal(err)
	}
	return f.request("POST", path+"/query", string(body), "stranger")
}

func TestServerSQLRejectsFilesystemAndEngineOperations(t *testing.T) {
	f, path := newSQLSecurityFixture(t)
	dir := t.TempDir() // Deliberately outside the server database directory.
	victimPath := filepath.Join(dir, "victim.db")
	victim, err := sql.Open("sqlite", victimPath)
	if err != nil {
		t.Fatal(err)
	}
	defer victim.Close()
	if _, err := victim.Exec(`CREATE TABLE users(id INTEGER PRIMARY KEY, role TEXT); INSERT INTO users VALUES(1, 'user')`); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(dir, "outside.db")
	sentinelPath := filepath.Join(dir, "sentinel.txt")
	if err := os.WriteFile(sentinelPath, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	attach := "ATTACH DATABASE " + sqlLiteral(victimPath) + " AS panel; UPDATE panel.users SET role='admin' WHERE id=1;"
	cases := []string{
		attach,
		"/* leading */ " + attach,
		"-- leading ATTACH\n" + attach,
		"SELECT 1; " + attach,
		"CREATE TABLE should_not_exist(id); /* mid */ " + attach,
		"; -- empty statement\n AtTaCh /* mid */ DATABASE " + sqlLiteral(victimPath) + " AS panel; UPDATE panel.users SET role='admin' WHERE id=1;",
		"SELECT 'ATTACH;'; /* safe prefix */ " + attach,
		"VACUUM INTO " + sqlLiteral(outputPath),
		"SELECT 1; -- mid\n VACUUM main INTO " + sqlLiteral(outputPath),
		"VACUUM INTO " + sqlLiteral(sentinelPath),
		"DETACH DATABASE main",
		"PRAGMA writable_schema=ON",
		"SELECT 1; /* mid */ PRAGMA journal_mode=OFF",
		"PRAGMA temp_store_directory=" + sqlLiteral(dir),
		"SELECT load_extension(" + sqlLiteral(sentinelPath) + ")",
		"SELECT 1; SELECT LoAd_ExTeNsIoN /* mid */ ('missing')",
		`SELECT "load_extension"('missing')`,
		"SELECT `load_extension`('missing')",
		`SELECT [load_extension]('missing')`,
		`SELECT 'load_extension'('missing')`,
		`SELECT * FROM pragma_table_info('users')`,
		`SELECT * FROM "pragma_table_info"('users')`,
		"SELECT writefile(" + sqlLiteral(sentinelPath) + ", 'changed')",
		"SELECT readfile(" + sqlLiteral(sentinelPath) + ")",
		"CREATE VIRTUAL TABLE external USING zipfile(" + sqlLiteral(outputPath) + ")",
		"CREATE VIEW delayed AS SELECT load_extension('missing')",
	}
	for i, query := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			w := requestSecuritySQL(t, f, path, query)
			if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "not permitted") {
				t.Fatalf("expected policy rejection: %d %s; SQL: %s", w.Code, w.Body.String(), query)
			}
			var role string
			if err := victim.QueryRow("SELECT role FROM users WHERE id=1").Scan(&role); err != nil || role != "user" {
				t.Fatalf("external role changed: %q, %v", role, err)
			}
			if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
				t.Fatalf("outside output exists or stat failed: %v", err)
			}
			if got, err := os.ReadFile(sentinelPath); err != nil || string(got) != "unchanged" {
				t.Fatalf("external sentinel changed: %q, %v", got, err)
			}
		})
	}
	w := requestSecuritySQL(t, f, path, "SELECT count(*) FROM sqlite_master WHERE name='should_not_exist'")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"rows":[[0]]`) {
		t.Fatalf("batch prefix executed before rejection: %d %s", w.Code, w.Body.String())
	}
}

func TestServerSQLOrdinaryQueriesAndLimits(t *testing.T) {
	f, path := newSQLSecurityFixture(t)
	for _, query := range []string{
		`CREATE TABLE "ATTACH"("PRAGMA" TEXT, [VACUUM] INTEGER)`,
		"INSERT INTO \"ATTACH\" VALUES('ATTACH DETACH VACUUM PRAGMA load_extension; -- /*', 1); INSERT INTO \"ATTACH\" VALUES('it''s ordinary data', 2)",
		"/* VACUUM INTO 'outside' */ UPDATE \"ATTACH\" SET [VACUUM]=3 WHERE [VACUUM]=2 -- ATTACH",
		`CREATE INDEX ordinary_index ON "ATTACH"([VACUUM]); CREATE VIEW ordinary_view AS SELECT * FROM "ATTACH"`,
		`CREATE TRIGGER ordinary_trigger AFTER INSERT ON "ATTACH" BEGIN UPDATE "ATTACH" SET [VACUUM]=4 WHERE [VACUUM]=5; END`,
	} {
		if w := requestSecuritySQL(t, f, path, query); w.Code != 200 {
			t.Fatalf("ordinary SQL failed: %s: %d %s", query, w.Code, w.Body.String())
		}
	}
	for _, query := range []string{
		`SELECT "PRAGMA", [VACUUM] FROM "ATTACH"`,
		"-- ATTACH\n /* PRAGMA */ SELECT 'VACUUM; ATTACH', 'load_extension', 'it''s PRAGMA', \"PRAGMA\" FROM \"ATTACH\"; -- trailing ;",
		`WITH x AS (SELECT * FROM "ATTACH") SELECT * FROM x`,
		`SELECT 'ATTACH' AS "DETACH", 1 AS ` + "`VACUUM`",
	} {
		w := requestSecuritySQL(t, f, path, query)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"mode":"rows"`) {
			t.Fatalf("ordinary query failed: %s: %d %s", query, w.Code, w.Body.String())
		}
	}
	// Trusted schema introspection must retain its own PRAGMA-capable connection.
	if w := f.request("GET", path+"/tables", "", "owner"); w.Code != 200 || !strings.Contains(w.Body.String(), "PRAGMA") {
		t.Fatalf("schema introspection failed: %d %s", w.Code, w.Body.String())
	}
	w := requestSecuritySQL(t, f, path, `WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<501) SELECT x FROM n`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"row_count":500`) || !strings.Contains(w.Body.String(), `"truncated":true`) {
		t.Fatalf("row limit changed: %d %s", w.Code, w.Body.String())
	}
	for _, query := range []string{
		fmt.Sprintf("SELECT zeroblob(%d)", maxServerSQLValueBytes+1),
		fmt.Sprintf("SELECT randomblob(%d)", maxServerSQLValueBytes+1),
		fmt.Sprintf("CREATE TABLE too_big AS SELECT zeroblob(%d) AS value", maxServerSQLValueBytes+1),
		"SELECT 1 -- " + strings.Repeat("x", maxServerSQLBytes),
	} {
		if w := requestSecuritySQL(t, f, path, query); w.Code != http.StatusBadRequest {
			t.Fatalf("oversized SQL/value accepted: %d %.200s", w.Code, w.Body.String())
		}
	}
	body := `{"sql":"SELECT 1"}` + strings.Repeat(" ", maxServerSQLRequestBytes)
	if w := f.request("POST", path+"/query", body, "stranger"); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized request accepted: %d %.200s", w.Code, w.Body.String())
	}
}

func TestServerSQLBoundsAggregateResponse(t *testing.T) {
	f, path := newSQLSecurityFixture(t)
	w := requestSecuritySQL(t, f, path, "WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<100) SELECT printf('%0100000d',x) AS payload FROM n")
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	var result struct {
		Truncated bool `json:"truncated"`
		RowCount  int  `json:"row_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Truncated || result.RowCount <= 0 || result.RowCount >= 100 || w.Body.Len() > 9<<20 {
		t.Fatal("aggregate SQL response was not bounded", result, w.Body.Len())
	}
}

func TestServerSQLSchemaDoesNotExecuteRecursiveViews(t *testing.T) {
	f, path := newSQLSecurityFixture(t)
	w := requestSecuritySQL(t, f, path, `CREATE VIEW runaway AS WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n) SELECT x FROM n`)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	start := time.Now()
	w = f.request("GET", path+"/tables", "", "stranger")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "runaway") {
		t.Fatal(w.Code, w.Body.String())
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("schema inspection executed an unbounded view")
	}
}
