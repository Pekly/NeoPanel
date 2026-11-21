package database

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// NodeIdlePolicy is desired policy for a remote node, or effective host policy for local.
type NodeIdlePolicy struct {
	Enabled        bool `json:"enabled"`
	TimeoutMinutes int  `json:"timeout_minutes"`
}

func DefaultNodeIdlePolicy() NodeIdlePolicy { return NodeIdlePolicy{TimeoutMinutes: 15} }

func ValidateNodeIdlePolicy(p NodeIdlePolicy) error {
	if p.TimeoutMinutes < 1 || p.TimeoutMinutes > 1440 {
		return fmt.Errorf("timeout_minutes must be between 1 and 1440")
	}
	return nil
}

func normalizeStoredIdlePolicy(p NodeIdlePolicy) (NodeIdlePolicy, error) {
	if p.TimeoutMinutes == 0 {
		p.TimeoutMinutes = 15
	}
	if err := ValidateNodeIdlePolicy(p); err != nil {
		return DefaultNodeIdlePolicy(), err
	}
	return p, nil
}

func decodeStoredIdlePolicy(raw string) (NodeIdlePolicy, error) {
	p := DefaultNodeIdlePolicy()
	d := json.NewDecoder(bytes.NewBufferString(raw))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return p, fmt.Errorf("invalid stored idle policy")
	}
	seen := map[string]bool{}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return DefaultNodeIdlePolicy(), err
		}
		key, ok := token.(string)
		if !ok || seen[key] || (key != "enabled" && key != "timeout_minutes") {
			return DefaultNodeIdlePolicy(), fmt.Errorf("invalid stored idle policy field")
		}
		seen[key] = true
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return DefaultNodeIdlePolicy(), err
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return DefaultNodeIdlePolicy(), fmt.Errorf("null idle policy field")
		}
		var dest any = &p.Enabled
		if key == "timeout_minutes" {
			dest = &p.TimeoutMinutes
		}
		if err := json.Unmarshal(value, dest); err != nil {
			return DefaultNodeIdlePolicy(), err
		}
	}
	if token, err = d.Token(); err != nil || token != json.Delim('}') || len(seen) != 2 || d.Decode(new(any)) != io.EOF {
		return DefaultNodeIdlePolicy(), fmt.Errorf("invalid stored idle policy")
	}
	return normalizeStoredIdlePolicy(p)
}

func (db *DB) GetNodeIdlePolicy(nodeID string) (NodeIdlePolicy, error) {
	if nodeID == "local" {
		raw, err := db.GetPanelSetting("node_idle_policy:local")
		if err == nil {
			return decodeStoredIdlePolicy(raw)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return DefaultNodeIdlePolicy(), err
		}
	}
	var enabled, minutes int
	err := db.conn.QueryRow(`SELECT COALESCE(is_free, 0), COALESCE(idle_timeout_min, 0) FROM nodes WHERE id = ?`, nodeID).Scan(&enabled, &minutes)
	if nodeID == "local" && errors.Is(err, sql.ErrNoRows) {
		return DefaultNodeIdlePolicy(), nil
	}
	if err != nil {
		return DefaultNodeIdlePolicy(), err
	}
	// IsFree is the sole enable authority, not a positive legacy timeout.
	return normalizeStoredIdlePolicy(NodeIdlePolicy{Enabled: enabled == 1, TimeoutMinutes: minutes})
}

func (db *DB) SetNodeIdlePolicy(nodeID string, p NodeIdlePolicy) error {
	if err := ValidateNodeIdlePolicy(p); err != nil {
		return err
	}
	if nodeID == "local" {
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}
		return db.SetPanelSetting("node_idle_policy:local", string(data))
	}
	enabled := 0
	if p.Enabled {
		enabled = 1
	}
	result, err := db.conn.Exec(`UPDATE nodes SET is_free = ?, idle_timeout_min = ?, updated_at = ? WHERE id = ?`, enabled, p.TimeoutMinutes, time.Now().UTC(), nodeID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return err
}
