package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// Reuse the durable settings store, one record per server/filename, not a growing
// history of every digest. Length-delimited JSON prevents ambiguous key joins.
func pluginSourceKey(serverID, filename string) string {
	b, _ := json.Marshal([]string{serverID, filename})
	h := sha256.Sum256(b)
	return "plugin_source_v1_" + hex.EncodeToString(h[:])
}

type pluginSource struct {
	SHA512    string `json:"sha512"`
	ProjectID string `json:"project_id"`
}

func (db *DB) ConfirmPluginSource(serverID, filename, sha512, projectID string) error {
	b, err := json.Marshal(pluginSource{sha512, projectID})
	if err != nil {
		return err
	}
	return db.SetPanelSetting(pluginSourceKey(serverID, filename), string(b))
}

// A changed file has no association, even when its filename is unchanged.
func (db *DB) ConfirmedPluginSource(serverID, filename, sha512 string) (string, error) {
	raw, err := db.GetPanelSetting(pluginSourceKey(serverID, filename))
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var source pluginSource
	if err := json.Unmarshal([]byte(raw), &source); err != nil {
		return "", err
	}
	if source.SHA512 != sha512 {
		return "", nil
	}
	return source.ProjectID, nil
}
