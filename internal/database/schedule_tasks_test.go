package database

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestScheduleTasksPersistenceLegacyAndAtomicUpdates(t *testing.T) {
	db, err := InitDB(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.MigrateScheduleTasks(); err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateScheduleTasks(); err != nil {
		t.Fatal("migration not idempotent", err)
	}
	if err := db.CreateServer(&Server{ID: "alpha", Name: "alpha", Type: "paper", Version: "1.21", JarFile: "server.jar"}); err != nil {
		t.Fatal(err)
	}
	s := &ServerSchedule{ID: "legacy", ServerID: "alpha", Name: "old", CronExpr: "* * * * *", Action: "backup", IsActive: true}
	if err := db.CreateServerSchedule(s); err != nil {
		t.Fatal(err)
	}
	tasks, err := db.ListScheduleTasks(s)
	if err != nil || len(tasks) != 1 || tasks[0].Order != 1 || tasks[0].Action != "backup" {
		t.Fatal(tasks, err)
	}
	tasks = []ScheduleTask{{Action: "command", Payload: "say hello", TimeOffset: 2, ContinueOnFailure: true}, {Action: "backup"}}
	if err := db.SaveScheduleWithTasks(s, tasks, false); err != nil {
		t.Fatal(err)
	}
	before, err := db.ListScheduleTasks(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Name = "renamed"
	if err := db.SaveScheduleWithTasks(s, nil, false); err != nil {
		t.Fatal(err)
	}
	after, err := db.ListScheduleTasks(s)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("omitted tasks replaced", before, after, err)
	}
	// Invalid task update cannot partially update the schedule itself.
	s.Name = "must not persist"
	if err := db.SaveScheduleWithTasks(s, []ScheduleTask{}, false); err == nil {
		t.Fatal("empty tasks accepted")
	}
	got, err := db.GetServerSchedule("alpha", s.ID)
	if err != nil || got.Name != "renamed" {
		t.Fatal(got, err)
	}
	now := time.Now().UTC()
	snapshot, claimedTasks, claimed, err := db.ClaimScheduleWithTasks(s.ID, now, true)
	if err != nil || !claimed || snapshot.Name != "renamed" || !reflect.DeepEqual(claimedTasks, before) {
		t.Fatal(snapshot, claimedTasks, claimed, err)
	}
	if _, _, claimed, err := db.ClaimScheduleWithTasks(s.ID, now, false); err != nil || claimed {
		t.Fatal("overlap", claimed, err)
	}
	if err := db.SetScheduleRunStatus(s.ID, "succeeded", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, claimed, err := db.ClaimScheduleWithTasks(s.ID, now, true); err != nil || claimed {
		t.Fatal("minute dedup lost", claimed, err)
	}
	if err := db.DeleteServerSchedule("alpha", s.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.conn.QueryRow(`SELECT count(*) FROM schedule_tasks WHERE schedule_id=?`, s.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("orphan tasks", count, err)
	}
}
