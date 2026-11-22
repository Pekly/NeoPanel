package runner

import (
	"github.com/Pekly/NeoPanel/internal/database"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDisableStopsNewScheduleSteps(t *testing.T) {
	root := t.TempDir()
	db, err := database.InitDB(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := NewServerManager(filepath.Join(root, "servers"), db)
	defer m.ShutdownAll()
	s := &database.ServerSchedule{ID: "test", ServerID: "alpha", Action: "start", CronExpr: "* * * * *"}
	if err := db.SetFeaturePolicy("", database.FeaturePolicy{"schedules": false}); err != nil {
		t.Fatal(err)
	}
	if err := m.StartSchedule(s); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatal(err)
	}
	if err := m.executeScheduleAction(s); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatal(err)
	}
	if err := db.SetFeaturePolicy("", database.FeaturePolicy{"backups": false}); err != nil {
		t.Fatal(err)
	}
	s.Action = "backup"
	if err := m.executeScheduleAction(s); err == nil || !strings.Contains(err.Error(), "backups are disabled") {
		t.Fatal(err)
	}
}
