package runner

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestScheduleTasksOrderedDelaysAndFailurePolicy(t *testing.T) {
	for _, keepGoing := range []bool{false, true} {
		var events []string
		tasks := []database.ScheduleTask{
			{Action: "command", Payload: "first", TimeOffset: 1},
			{Action: "backup", TimeOffset: 2, ContinueOnFailure: keepGoing},
			{Action: "start", TimeOffset: 3},
		}
		err := executeScheduleTasks(tasks, func(d time.Duration) error { events = append(events, d.String()); return nil }, func(task database.ScheduleTask) error {
			events = append(events, task.Action)
			if task.Action == "backup" {
				return errors.New("disk full")
			}
			return nil
		})
		want := []string{"1s", "command", "2s", "backup"}
		if keepGoing {
			want = append(want, "3s", "start")
		}
		if !reflect.DeepEqual(events, want) {
			t.Fatal(events, want)
		}
		if err == nil || !strings.Contains(err.Error(), "task 2 (backup): disk full") {
			t.Fatal("failure was hidden", err)
		}
	}
}

func TestScheduleTasksCancelledWaitStopsEvenContinueSteps(t *testing.T) {
	tasks := []database.ScheduleTask{{Action: "backup", ContinueOnFailure: true}, {Action: "start"}}
	calls := 0
	err := executeScheduleTasks(tasks, func(time.Duration) error { return errors.New("cancelled") }, func(database.ScheduleTask) error { calls++; return nil })
	if err == nil || calls != 0 {
		t.Fatal(err, calls)
	}
	m := NewServerManager(t.TempDir(), nil)
	m.ShutdownAll()
	if err := m.waitScheduleDelay(time.Hour); err == nil {
		t.Fatal("shutdown did not cancel")
	}
}

func TestMultiStepScheduleSnapshotAndHonestCompletion(t *testing.T) {
	root := t.TempDir()
	db, err := database.InitDB(filepath.Join(root, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.MigrateScheduleTasks(); err != nil {
		t.Fatal(err)
	}
	srv := &database.Server{ID: "alpha", Name: "alpha", Type: "paper", Version: "1.21", JarFile: "server.jar"}
	if err := db.CreateServer(srv); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(filepath.Join(root, "servers"), db)
	defer m.ShutdownAll()
	os.MkdirAll(m.GetWorkingDir(srv.ID), 0755)
	os.WriteFile(filepath.Join(m.GetWorkingDir(srv.ID), "world"), []byte("world"), 0600)
	s := &database.ServerSchedule{ID: "multi", ServerID: srv.ID, Name: "multi", CronExpr: "* * * * *", IsActive: true}
	tasks := []database.ScheduleTask{{Action: "command", Payload: "offline fails", TimeOffset: 1, ContinueOnFailure: true}, {Action: "backup"}}
	if err := db.SaveScheduleWithTasks(s, tasks, true); err != nil {
		t.Fatal(err)
	}
	if err := m.StartSchedule(s); err != nil {
		t.Fatal(err)
	}
	if err := m.StartSchedule(s); !errors.Is(err, ErrScheduleRunning) {
		t.Fatal("overlap accepted", err)
	}
	if err := db.DeleteServerSchedule(srv.ID, s.ID); err == nil {
		t.Fatal("active schedule deleted")
	}
	// Editing while the first step waits affects only the next invocation.
	if err := db.SaveScheduleWithTasks(s, []database.ScheduleTask{{Action: "stop"}}, false); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		result, err := db.GetServerSchedule(srv.ID, s.ID)
		if err != nil {
			t.Fatal(err)
		}
		if result.LastStatus == "failed" {
			if result.LastFinishedAt == nil || !strings.Contains(result.LastError, "task 1 (command)") {
				t.Fatal(result)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("run did not finish", result)
		}
		time.Sleep(10 * time.Millisecond)
	}
	backups, err := os.ReadDir(filepath.Join(root, "backups", srv.ID))
	if err != nil || len(backups) != 1 {
		t.Fatal("snapshot backup step not executed", backups, err)
	}
}
