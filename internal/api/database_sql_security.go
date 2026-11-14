package api

import (
	"database/sql"
	"fmt"
	"strings"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	maxServerSQLRequestBytes = 2 << 20
	maxServerSQLBytes        = 256 << 10
	maxServerSQLValueBytes   = 1 << 20
)

// restrictServerSQLConn must run on the very connection used to execute user SQL,
// not on a transient connection borrowed from the pool. Trusted introspection
// uses separate connections and can still issue PRAGMAs.
func restrictServerSQLConn(conn *sql.Conn) error {
	for _, limit := range []struct{ id, value int }{
		{sqlite3.SQLITE_LIMIT_ATTACHED, 0},
		{sqlite3.SQLITE_LIMIT_SQL_LENGTH, maxServerSQLBytes},
		{sqlite3.SQLITE_LIMIT_LENGTH, maxServerSQLValueBytes},
	} {
		if _, err := sqlite.Limit(conn, limit.id, limit.value); err != nil {
			return err
		}
	}
	return nil
}

// inspectServerSQL scans the entire batch before any statement executes. The
// attachment limit also blocks VACUUM INTO in the current driver, but explicitly
// deny it regardless. The first keyword only selects the response format.
func inspectServerSQL(query string) (keyword string, batch bool, err error) {
	if len(query) > maxServerSQLBytes {
		return "", false, fmt.Errorf("SQL exceeds %d bytes", maxServerSQLBytes)
	}
	if strings.IndexByte(query, 0) >= 0 {
		return "", false, fmt.Errorf("SQL contains a NUL byte")
	}
	ended, seen := false, false
	previous, pendingFunction := "", ""
	for pos := 0; pos < len(query); {
		var token string
		var quoted bool
		token, quoted, pos, err = nextServerSQLToken(query, pos)
		if err != nil {
			return "", false, err
		}
		if token == "" && !quoted {
			break
		}
		upper := strings.ToUpper(token)
		if pendingFunction != "" && token == "(" && !quoted {
			return "", false, fmt.Errorf("SQL function %s is not permitted", pendingFunction)
		}
		pendingFunction = ""
		if unsafeServerSQLFunction(upper) {
			// SQLite permits quoted function names, including single quotes.
			// Quoted data and ordinary quoted identifiers remain legal.
			pendingFunction = upper
			if previous == "USING" {
				return "", false, fmt.Errorf("SQL module %s is not permitted", upper)
			}
		}
		if !quoted {
			switch upper {
			case "ATTACH", "DETACH", "VACUUM", "PRAGMA":
				return "", false, fmt.Errorf("SQL operation %s is not permitted", upper)
			}
		}
		if !quoted && token == ";" {
			ended = seen
		} else {
			if ended {
				batch = true
			}
			if !seen {
				keyword = upper
			}
			seen = true
		}
		previous = upper
	}
	return keyword, batch, nil
}

func unsafeServerSQLFunction(name string) bool {
	// Include extension file-I/O functions/modules even when the current driver
	// does not register them. PRAGMA table-valued functions must not bypass the
	// prohibition on user-controlled engine settings/introspection.
	switch name {
	case "LOAD_EXTENSION", "READFILE", "WRITEFILE", "EDIT", "FSDIR", "ZIPFILE":
		return true
	}
	return strings.HasPrefix(name, "PRAGMA_")
}

// nextServerSQLToken follows SQLite quoting/comment rules, not SQL substrings.
// Quoted tokens are decoded solely to recognize dangerous callable names; their
// contents are never interpreted as statement keywords or batch delimiters.
func nextServerSQLToken(s string, pos int) (token string, quoted bool, next int, err error) {
	for pos < len(s) {
		switch s[pos] {
		case ' ', '\t', '\r', '\n', '\f':
			pos++
			continue
		}
		if strings.HasPrefix(s[pos:], "\xef\xbb\xbf") {
			pos += 3
			continue
		}
		if strings.HasPrefix(s[pos:], "--") {
			if end := strings.IndexByte(s[pos+2:], '\n'); end >= 0 {
				pos += end + 3
				continue
			}
			return "", false, len(s), nil
		}
		if strings.HasPrefix(s[pos:], "/*") {
			if end := strings.Index(s[pos+2:], "*/"); end >= 0 {
				pos += end + 4
				continue
			}
			return "", false, 0, fmt.Errorf("Unterminated SQL comment")
		}
		break
	}
	if pos == len(s) {
		return "", false, pos, nil
	}
	start := pos
	switch quote := s[pos]; quote {
	case '\'', '"', '`', '[':
		endQuote := quote
		if quote == '[' {
			endQuote = ']'
		}
		var value strings.Builder
		for pos++; pos < len(s); pos++ {
			if s[pos] == endQuote {
				if quote != '[' && pos+1 < len(s) && s[pos+1] == endQuote {
					value.WriteByte(endQuote)
					pos++
					continue
				}
				return value.String(), true, pos + 1, nil
			}
			value.WriteByte(s[pos])
		}
		return "", false, 0, fmt.Errorf("Unterminated SQL quoted token")
	}
	if serverSQLIdentifierByte(s[pos]) {
		for pos < len(s) && serverSQLIdentifierByte(s[pos]) {
			pos++
		}
		return s[start:pos], false, pos, nil
	}
	return s[pos : pos+1], false, pos + 1, nil
}

func serverSQLIdentifierByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '$' || c >= 0x80
}
