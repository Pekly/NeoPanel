package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// FeatureKeys is the allowlist for optional panel tools. Authentication, access
// control and basic file recovery are deliberately not optional.
var FeatureKeys = []string{"backups", "schedules", "sftp", "sftp_import", "plugins", "databases", "players", "proxy_settings", "extra_ports", "software", "subdomains", "countdowns", "world_conversion"}

type FeaturePolicy map[string]bool

func ValidFeature(key string) bool {
	for _, name := range FeatureKeys {
		if key == name {
			return true
		}
	}
	return false
}

func ValidateFeaturePolicy(p FeaturePolicy) error {
	for key := range p {
		if !ValidFeature(key) {
			return fmt.Errorf("unknown feature: %s", key)
		}
	}
	return nil
}

func featureSetting(serverID string) string {
	if serverID == "" {
		return "features:global"
	}
	return "features:server:" + serverID
}

func (db *DB) GetFeaturePolicy(serverID string) (FeaturePolicy, error) {
	raw, err := db.GetPanelSetting(featureSetting(serverID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	p := FeaturePolicy{}
	if raw != "" {
		if err = json.Unmarshal([]byte(raw), &p); err != nil {
			return nil, err
		}
	}
	if p == nil {
		p = FeaturePolicy{}
	}
	return p, ValidateFeaturePolicy(p)
}

// SetFeaturePolicy replaces a sparse policy. Missing values inherit; a global
// false is a master disable and cannot be overridden by a server.
func (db *DB) SetFeaturePolicy(serverID string, p FeaturePolicy) error {
	if err := ValidateFeaturePolicy(p); err != nil {
		return err
	}
	if p == nil {
		p = FeaturePolicy{}
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return db.SetPanelSetting(featureSetting(serverID), string(data))
}

func (db *DB) EffectiveFeatures(serverID string) (FeaturePolicy, error) {
	global, err := db.GetFeaturePolicy("")
	if err != nil {
		return nil, err
	}
	local := FeaturePolicy{}
	if serverID != "" {
		local, err = db.GetFeaturePolicy(serverID)
		if err != nil {
			return nil, err
		}
	}
	effective := FeaturePolicy{}
	for _, key := range FeatureKeys {
		g, set := global[key]
		if !set {
			g = true
		}
		s, set := local[key]
		if !set {
			s = true
		}
		effective[key] = g && s
	}
	return effective, nil
}

// FeatureEnabled fails closed on unreadable policy, including background uses.
func (db *DB) FeatureEnabled(serverID, key string) bool {
	if db == nil || !ValidFeature(key) {
		return false
	}
	p, err := db.EffectiveFeatures(serverID)
	return err == nil && p[key]
}
