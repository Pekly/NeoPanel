package database

import (
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNodeIdlePolicyPersistenceIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defaults := DefaultNodeIdlePolicy()
	if defaults.Enabled || defaults.TimeoutMinutes != 15 {
		t.Fatal(defaults)
	}
	if p, err := db.GetNodeIdlePolicy("local"); err != nil || p != defaults {
		t.Fatal(p, err)
	}
	node := &ClusterNode{ID: "remote", Name: "keep", BaseURL: "https://worker", APIToken: "secret", FQDN: "fqdn", Location: "location", AllowUserRAM: true, AllowUserPorts: true, AllowUserJava: false, IsolationEngine: "docker", CPULimitPercent: 75, TasksMax: 99}
	if err := db.CreateNode(node); err != nil {
		t.Fatal(err)
	}
	before, _ := db.GetNode("remote")
	policy := NodeIdlePolicy{Enabled: true, TimeoutMinutes: 42}
	if err := db.SetNodeIdlePolicy("remote", policy); err != nil {
		t.Fatal(err)
	}
	after, _ := db.GetNode("remote")
	before.IsFree, before.IdleTimeoutMin, before.UpdatedAt = after.IsFree, after.IdleTimeoutMin, after.UpdatedAt
	if !reflect.DeepEqual(before, after) {
		t.Fatal("unrelated columns changed", before, after)
	}
	if p, err := db.GetNodeIdlePolicy("remote"); err != nil || p != policy {
		t.Fatal(p, err)
	}
	if p, _ := db.GetNodeIdlePolicy("local"); p != defaults {
		t.Fatal("remote changed local", p)
	}
	if err := db.SetNodeIdlePolicy("local", policy); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetNode("local"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("synthetic local node", err)
	}
	if err := db.SetNodeIdlePolicy("missing", policy); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"local", "remote"} {
		if p, err := db.GetNodeIdlePolicy(id); err != nil || p != policy {
			t.Fatal(id, p, err)
		}
	}
	for _, minutes := range []int{-1, 0, 1441} {
		bad := NodeIdlePolicy{Enabled: true, TimeoutMinutes: minutes}
		if ValidateNodeIdlePolicy(bad) == nil || db.SetNodeIdlePolicy("local", bad) == nil {
			t.Fatal("accepted invalid", bad)
		}
	}
	for _, minutes := range []int{1, 1440} {
		if err := ValidateNodeIdlePolicy(NodeIdlePolicy{TimeoutMinutes: minutes}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNodeIdlePolicyLegacyAndCorruption(t *testing.T) {
	db, err := InitDB(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"local", "remote"} {
		node := &ClusterNode{ID: id, Name: id}
		if err := db.CreateNode(node); err != nil {
			t.Fatal(err)
		}
		for _, enabled := range []bool{false, true} {
			for _, minutes := range []int{0, 15, 60, -1, 1441} {
				node.IsFree, node.IdleTimeoutMin = enabled, minutes
				if err := db.UpdateNode(node); err != nil {
					t.Fatal(err)
				}
				p, err := db.GetNodeIdlePolicy(id)
				if minutes < 0 || minutes > 1440 {
					if err == nil || p.Enabled {
						t.Fatal("unsafe legacy", p, err)
					}
					continue
				}
				want := minutes
				if want == 0 {
					want = 15
				}
				if err != nil || p.Enabled != enabled || p.TimeoutMinutes != want {
					t.Fatal(p, err)
				}
			}
		}
	}
	for _, raw := range []string{`{"enabled":true`, `null`, `{}`, `{"enabled":true}`, `{"enabled":true,"timeout_minutes":-1}`, `{"enabled":true,"timeout_minutes":1441}`, `{"enabled":true,"timeout_minutes":null}`, `{"enabled":true,"enabled":false,"timeout_minutes":15}`, `{"Enabled":true,"timeout_minutes":15}`} {
		if err := db.SetPanelSetting("node_idle_policy:local", raw); err != nil {
			t.Fatal(err)
		}
		if p, err := db.GetNodeIdlePolicy("local"); err == nil || p != DefaultNodeIdlePolicy() {
			t.Fatal(raw, p, err)
		}
	}
	if err := db.SetPanelSetting("node_idle_policy:local", `{"enabled":false,"timeout_minutes":0}`); err != nil {
		t.Fatal(err)
	}
	if p, err := db.GetNodeIdlePolicy("local"); err != nil || p != DefaultNodeIdlePolicy() {
		t.Fatal(p, err)
	}
}
