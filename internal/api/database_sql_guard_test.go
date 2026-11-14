package api

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func TestServerSQLConnectionLimits(t *testing.T) {
	pool, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "server.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ctx := context.Background()
	conn, err := pool.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := restrictServerSQLConn(conn); err != nil {
		t.Fatal(err)
	}
	for _, limit := range []struct{ id, want int }{
		{sqlite3.SQLITE_LIMIT_ATTACHED, 0},
		{sqlite3.SQLITE_LIMIT_SQL_LENGTH, maxServerSQLBytes},
		{sqlite3.SQLITE_LIMIT_LENGTH, maxServerSQLValueBytes},
	} {
		if got, err := sqlite.Limit(conn, limit.id, -1); err != nil || got != limit.want {
			t.Fatalf("limit %d = %d, want %d: %v", limit.id, got, limit.want, err)
		}
	}
	// Bypass the lexical guard: the execution connection itself must reject ATTACH.
	if _, err := conn.ExecContext(ctx, `ATTACH DATABASE ':memory:' AS forbidden`); err == nil || !strings.Contains(err.Error(), "too many attached databases") {
		t.Fatalf("actual execution connection allows ATTACH: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "SELECT 1 -- "+strings.Repeat("x", maxServerSQLBytes)); err == nil {
		t.Fatal("SQLite SQL length limit not applied")
	}
	var value []byte
	if err := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT zeroblob(%d)", maxServerSQLValueBytes+1)).Scan(&value); err == nil {
		t.Fatal("SQLite value length limit not applied")
	}
	// A second simultaneously borrowed connection is intentionally unrestricted:
	// proves why using sql.DB.Exec after limiting a transient sql.Conn is unsafe.
	other, err := pool.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if _, err := other.ExecContext(ctx, `ATTACH DATABASE ':memory:' AS allowed`); err != nil {
		t.Fatalf("second connection unexpectedly restricted: %v", err)
	}
	// SQLITE_LIMIT_ATTACHED=0 also blocks VACUUM INTO in the current driver.
	output := filepath.Join(t.TempDir(), "vacuum.db")
	query := "VACUUM INTO '" + strings.ReplaceAll(output, "'", "''") + "'"
	if _, err := conn.ExecContext(ctx, query); err == nil || !strings.Contains(err.Error(), "too many attached databases") {
		t.Fatalf("restricted connection allows VACUUM INTO: %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("restricted VACUUM produced output: %v", err)
	}
	if _, err := other.ExecContext(ctx, query); err != nil {
		t.Fatalf("unrestricted VACUUM control failed: %v", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected unrestricted VACUUM output: %v", err)
	}
}

func TestInspectServerSQLLexing(t *testing.T) {
	for _, query := range []string{
		"SELECT 'ATTACH; PRAGMA; load_extension(1)'",
		"SELECT 1 /* VACUUM */; -- ATTACH\rstill comment",
		"SELECT \"ATTACH\", [VACUUM], `PRAGMA`, 'it''s DETACH'",
		"SELECT 'load_extension', 'writefile', 'pragma_table_info'",
		"SELECT /* outer /* nested-looking */ 1",
	} {
		if _, _, err := inspectServerSQL(query); err != nil {
			t.Errorf("safe lexing %q: %v", query, err)
		}
	}
	for _, query := range []string{
		"\xef\xbb\xbf ATTACH ':memory:' AS x",
		"SELECT 1; /* outer /* nested-looking */ VACUUM INTO 'x'",
		"SELECT 'it''s safe'; PRAGMA writable_schema=1",
		"SELECT 1; SELECT \"load_extension\" /* comment */ ('x')",
		"SELECT 'unterminated",
		"SELECT 1 /* unterminated",
		"SELECT 1\x00; ATTACH ':memory:' AS x",
	} {
		if _, _, err := inspectServerSQL(query); err == nil {
			t.Errorf("unsafe/malformed SQL accepted: %q", query)
		}
	}
}
