package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	_ "modernc.org/sqlite"
)

var validDBNameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{1,40}$`)

type DatabaseHandler struct {
	cfg *config.Config
	db  *database.DB
}

func NewDatabaseHandler(cfg *config.Config, db *database.DB) *DatabaseHandler {
	return &DatabaseHandler{
		cfg: cfg,
		db:  db,
	}
}

type ServerDatabaseView struct {
	*database.ServerDatabase
	SizeBytes  int64  `json:"size_bytes"`
	TableCount int    `json:"table_count"`
	JDBCURL    string `json:"jdbc_url"`
	SQLiteDSN  string `json:"sqlite_dsn"`
}

type DBColumnInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	NotNull    bool   `json:"not_null"`
	PrimaryKey bool   `json:"primary_key"`
	DefaultVal string `json:"default_val"`
}

type DBTableInfo struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	RowCount int64          `json:"row_count"`
	Columns  []DBColumnInfo `json:"columns"`
	SQL      string         `json:"sql"`
}

func (h *DatabaseHandler) serverDBDir(serverID string) string {
	base := "./data"
	if h.cfg != nil && strings.TrimSpace(h.cfg.Paths.DataDir) != "" {
		base = h.cfg.Paths.DataDir
	}
	return filepath.Join(base, "databases", serverID)
}

func (h *DatabaseHandler) openServerSQLite(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=journal_mode(WAL)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	return conn, nil
}

func detectDBCapabilities() map[string]interface{} {
	mariaPath, err := exec.LookPath("mariadb")
	if err != nil {
		mariaPath, _ = exec.LookPath("mysql")
	}
	dumpPath, err := exec.LookPath("mariadb-dump")
	if err != nil {
		dumpPath, _ = exec.LookPath("mysqldump")
	}
	return map[string]interface{}{
		"sqlite":        true,
		"mariadb_cli":   mariaPath != "",
		"mariadb_dump":  dumpPath != "",
		"default":       "sqlite",
		"engines_label": "Built-in SQLite (pure Go)",
	}
}

func randomPasswordHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func quoteSQLIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

func enrichDBView(d *database.ServerDatabase) ServerDatabaseView {
	view := ServerDatabaseView{
		ServerDatabase: d,
		JDBCURL:        fmt.Sprintf("jdbc:sqlite:%s", filepath.ToSlash(d.FilePath)),
		SQLiteDSN:      d.FilePath,
	}
	if d.FilePath != "" {
		if st, err := os.Stat(d.FilePath); err == nil {
			view.SizeBytes = st.Size()
		}
		if walSt, err := os.Stat(d.FilePath + "-wal"); err == nil {
			view.SizeBytes += walSt.Size()
		}
	}
	return view
}

func (h *DatabaseHandler) countTables(dbPath string) int {
	if dbPath == "" {
		return 0
	}
	if _, err := os.Stat(dbPath); err != nil {
		return 0
	}
	conn, err := h.openServerSQLite(dbPath)
	if err != nil {
		return 0
	}
	defer conn.Close()

	var count int
	_ = conn.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%'`).Scan(&count)
	return count
}

func (h *DatabaseHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if _, err := h.db.GetServer(serverID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	items, err := h.db.ListServerDatabases(serverID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	out := make([]ServerDatabaseView, len(items))
	for i, d := range items {
		v := enrichDBView(d)
		v.TableCount = h.countTables(d.FilePath)
		out[i] = v
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"databases":    out,
		"capabilities": detectDBCapabilities(),
	})
}

func (h *DatabaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if _, err := h.db.GetServer(serverID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		Engine         string `json:"engine"`
		MaxConnections int    `json:"max_connections"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if !validDBNameRe.MatchString(req.Name) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Database name must start with a letter and contain 2-41 alphanumeric or underscore characters",
		})
		return
	}

	engine := strings.ToLower(strings.TrimSpace(req.Engine))
	if engine == "" || engine == "sqlite" {
		engine = "sqlite"
	} else {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Unsupported or unconfigured engine '%s': built-in pure-Go SQLite is active on this node", engine),
		})
		return
	}

	if req.MaxConnections <= 0 {
		req.MaxConnections = 25
	} else if req.MaxConnections > 500 {
		req.MaxConnections = 500
	}

	dbDir := h.serverDBDir(serverID)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to prepare database directory: " + err.Error()})
		return
	}

	dbFile := filepath.Join(dbDir, req.Name+".db")
	if absPath, err := filepath.Abs(dbFile); err == nil {
		dbFile = absPath
	}

	conn, err := h.openServerSQLite(dbFile)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to initialize SQLite database: " + err.Error()})
		return
	}
	_, initErr := conn.Exec(`PRAGMA journal_mode=WAL; PRAGMA user_version=1;`)
	_ = conn.Close()
	if initErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to initialize SQLite schema header: " + initErr.Error()})
		return
	}

	idBytes := make([]byte, 4)
	_, _ = rand.Read(idBytes)
	dbID := "db-" + hex.EncodeToString(idBytes)

	uBytes := make([]byte, 4)
	_, _ = rand.Read(uBytes)
	username := "u_" + hex.EncodeToString(uBytes)
	password := randomPasswordHex(12)

	record := &database.ServerDatabase{
		ID:             dbID,
		ServerID:       serverID,
		Name:           req.Name,
		Engine:         engine,
		Username:       username,
		Password:       password,
		Host:           "127.0.0.1",
		Port:           0,
		FilePath:       dbFile,
		MaxConnections: req.MaxConnections,
	}

	if err := h.db.CreateServerDatabase(record); err != nil {
		_ = os.Remove(dbFile)
		_ = os.Remove(dbFile + "-wal")
		_ = os.Remove(dbFile + "-shm")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "A database with this name already exists on this server"})
		return
	}

	logAudit(h.db, r, "db_create", fmt.Sprintf("Created %s database '%s' for %s", engine, req.Name, auditServerLabel(h.db, serverID)))

	view := enrichDBView(record)
	view.TableCount = 0
	writeJSON(w, http.StatusCreated, view)
}

func (h *DatabaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dbID := getURLParam(r, "dbId")

	record, err := h.db.GetServerDatabase(serverID, dbID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database not found"})
		return
	}

	if err := h.db.DeleteServerDatabase(serverID, dbID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete database record"})
		return
	}

	if record.FilePath != "" {
		_ = os.Remove(record.FilePath)
		_ = os.Remove(record.FilePath + "-wal")
		_ = os.Remove(record.FilePath + "-shm")
	}

	logAudit(h.db, r, "db_delete", fmt.Sprintf("Deleted database '%s' from %s", record.Name, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Database deleted successfully"})
}

func (h *DatabaseHandler) RotatePassword(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dbID := getURLParam(r, "dbId")

	record, err := h.db.GetServerDatabase(serverID, dbID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database not found"})
		return
	}

	newPassword := randomPasswordHex(12)
	if err := h.db.UpdateServerDatabasePassword(serverID, dbID, newPassword); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to rotate database password"})
		return
	}
	record.Password = newPassword

	logAudit(h.db, r, "db_rotate", fmt.Sprintf("Rotated credentials for database '%s' on %s", record.Name, auditServerLabel(h.db, serverID)))

	view := enrichDBView(record)
	view.TableCount = h.countTables(record.FilePath)
	writeJSON(w, http.StatusOK, view)
}

func (h *DatabaseHandler) Tables(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dbID := getURLParam(r, "dbId")

	record, err := h.db.GetServerDatabase(serverID, dbID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database not found"})
		return
	}

	conn, err := h.openServerSQLite(record.FilePath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to open SQLite database: " + err.Error()})
		return
	}
	defer conn.Close()

	rows, err := conn.Query(`SELECT name, type, COALESCE(sql, '') FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' ORDER BY type, name`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to inspect schema: " + err.Error()})
		return
	}
	defer rows.Close()

	var tables []DBTableInfo
	for rows.Next() {
		var t DBTableInfo
		if err := rows.Scan(&t.Name, &t.Type, &t.SQL); err != nil {
			continue
		}
		tables = append(tables, t)
	}
	_ = rows.Close()

	for i := range tables {
		qIdent := quoteSQLIdent(tables[i].Name)
		_ = conn.QueryRow("SELECT COUNT(*) FROM " + qIdent).Scan(&tables[i].RowCount)

		colRows, err := conn.Query("PRAGMA table_info(" + qIdent + ")")
		if err == nil {
			for colRows.Next() {
				var cid int
				var col DBColumnInfo
				var notNullInt, pkInt int
				var dflt sql.NullString
				if scanErr := colRows.Scan(&cid, &col.Name, &col.Type, &notNullInt, &dflt, &pkInt); scanErr == nil {
					col.NotNull = notNullInt == 1
					col.PrimaryKey = pkInt > 0
					if dflt.Valid {
						col.DefaultVal = dflt.String
					}
					tables[i].Columns = append(tables[i].Columns, col)
				}
			}
			_ = colRows.Close()
		}
		if tables[i].Columns == nil {
			tables[i].Columns = []DBColumnInfo{}
		}
	}

	if tables == nil {
		tables = []DBTableInfo{}
	}

	view := enrichDBView(record)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tables":     tables,
		"size_bytes": view.SizeBytes,
	})
}

func firstSQLKeyword(query string) string {
	s := strings.TrimSpace(query)
	for {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "--") {
			if idx := strings.IndexByte(s, '\n'); idx >= 0 {
				s = s[idx+1:]
				continue
			}
			return ""
		}
		if strings.HasPrefix(s, "/*") {
			if idx := strings.Index(s, "*/"); idx >= 0 {
				s = s[idx+2:]
				continue
			}
			return ""
		}
		break
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToUpper(fields[0])
}

func normalizeSQLValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return val
	}
}

func (h *DatabaseHandler) Query(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dbID := getURLParam(r, "dbId")

	record, err := h.db.GetServerDatabase(serverID, dbID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database not found"})
		return
	}

	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	query := strings.TrimSpace(req.SQL)
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SQL query cannot be empty"})
		return
	}

	conn, err := h.openServerSQLite(record.FilePath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to open SQLite database: " + err.Error()})
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	started := time.Now()
	kw := firstSQLKeyword(query)
	isQuery := kw == "SELECT" || kw == "PRAGMA" || kw == "EXPLAIN" || kw == "WITH" || kw == "VALUES"

	if isQuery && !strings.Contains(strings.TrimSuffix(query, ";"), ";") {
		rows, err := conn.QueryContext(ctx, query)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		const maxRows = 500
		resultRows := [][]interface{}{}
		truncated := false

		for rows.Next() {
			if len(resultRows) >= maxRows {
				truncated = true
				break
			}
			dest := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range dest {
				ptrs[i] = &dest[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			rowOut := make([]interface{}, len(cols))
			for i, v := range dest {
				rowOut[i] = normalizeSQLValue(v)
			}
			resultRows = append(resultRows, rowOut)
		}
		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		durMS := time.Since(started).Milliseconds()
		logAudit(h.db, r, "db_query", fmt.Sprintf("Executed SQL query on '%s' (%s): %s", record.Name, auditServerLabel(h.db, serverID), truncateAuditText(query, 90)))

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"mode":        "rows",
			"columns":     cols,
			"rows":        resultRows,
			"row_count":   len(resultRows),
			"truncated":   truncated,
			"duration_ms": durMS,
		})
		return
	}

	res, err := conn.ExecContext(ctx, query)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	durMS := time.Since(started).Milliseconds()

	logAudit(h.db, r, "db_query", fmt.Sprintf("Executed SQL statement on '%s' (%s): %s", record.Name, auditServerLabel(h.db, serverID), truncateAuditText(query, 90)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"mode":           "exec",
		"rows_affected":  affected,
		"last_insert_id": lastID,
		"duration_ms":    durMS,
	})
}

func sqlLiteral(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case bool:
		if val {
			return "1"
		}
		return "0"
	case []byte:
		return "X'" + hex.EncodeToString(val) + "'"
	case string:
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	case time.Time:
		return "'" + val.Format(time.RFC3339) + "'"
	default:
		return "'" + strings.ReplaceAll(fmt.Sprintf("%v", val), "'", "''") + "'"
	}
}

func (h *DatabaseHandler) Export(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dbID := getURLParam(r, "dbId")
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "db"
	}

	record, err := h.db.GetServerDatabase(serverID, dbID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database not found"})
		return
	}

	if format == "db" {
		// Flush WAL frames into the main .db file cleanly before reading
		if conn, err := h.openServerSQLite(record.FilePath); err == nil {
			_, _ = conn.Exec(`PRAGMA wal_checkpoint(TRUNCATE);`)
			_ = conn.Close()
		}

		f, err := os.Open(record.FilePath)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database file not found on disk"})
			return
		}
		defer f.Close()

		logAudit(h.db, r, "db_export", fmt.Sprintf("Downloaded SQLite binary '%s.db' from %s", record.Name, auditServerLabel(h.db, serverID)))

		w.Header().Set("Content-Type", "application/x-sqlite3")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.db"`, record.Name))
		_, _ = io.Copy(w, f)
		return
	}

	if format == "sql" {
		conn, err := h.openServerSQLite(record.FilePath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to open SQLite database: " + err.Error()})
			return
		}
		defer conn.Close()

		var b strings.Builder
		b.WriteString(fmt.Sprintf("-- NeoPanel SQLite Dump: %s (Server: %s)\n", record.Name, serverID))
		b.WriteString(fmt.Sprintf("-- Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))
		b.WriteString("PRAGMA foreign_keys=OFF;\nBEGIN TRANSACTION;\n\n")

		ddlRows, err := conn.Query(`SELECT name, type, sql FROM sqlite_master WHERE type IN ('table', 'index', 'trigger', 'view') AND sql IS NOT NULL AND name NOT LIKE 'sqlite_%' ORDER BY CASE type WHEN 'table' THEN 1 WHEN 'index' THEN 2 WHEN 'trigger' THEN 3 ELSE 4 END, name`)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to export schema: " + err.Error()})
			return
		}

		type ddlEntry struct {
			name string
			typ  string
			sql  string
		}
		var ddls []ddlEntry
		for ddlRows.Next() {
			var e ddlEntry
			if err := ddlRows.Scan(&e.name, &e.typ, &e.sql); err == nil {
				ddls = append(ddls, e)
			}
		}
		_ = ddlRows.Close()

		for _, e := range ddls {
			if e.typ == "table" {
				b.WriteString(e.sql + ";\n")
				qTbl := quoteSQLIdent(e.name)
				dataRows, err := conn.Query("SELECT * FROM " + qTbl)
				if err == nil {
					cols, _ := dataRows.Columns()
					for dataRows.Next() {
						dest := make([]interface{}, len(cols))
						ptrs := make([]interface{}, len(cols))
						for i := range dest {
							ptrs[i] = &dest[i]
						}
						if err := dataRows.Scan(ptrs...); err == nil {
							vals := make([]string, len(dest))
							for i, v := range dest {
								vals[i] = sqlLiteral(v)
							}
							b.WriteString(fmt.Sprintf("INSERT INTO %s VALUES (%s);\n", qTbl, strings.Join(vals, ", ")))
						}
					}
					_ = dataRows.Close()
				}
				b.WriteString("\n")
			}
		}

		for _, e := range ddls {
			if e.typ != "table" {
				b.WriteString(e.sql + ";\n")
			}
		}

		b.WriteString("\nCOMMIT;\n")

		logAudit(h.db, r, "db_export", fmt.Sprintf("Exported SQL dump '%s.sql' from %s", record.Name, auditServerLabel(h.db, serverID)))

		w.Header().Set("Content-Type", "application/sql; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.sql"`, record.Name))
		_, _ = io.WriteString(w, b.String())
		return
	}

	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported export format; use format=db or format=sql"})
}

