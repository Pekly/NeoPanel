package database

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCronValidationMatchingAndNext(t *testing.T) {
	for _, expr := range []string{"", "* * * *", "* * * * * *", "60 * * * *", "* 24 * * *", "* * 0 * *", "* * * 13 *", "* * * * 8", "*/0 * * * *", "1-5/0 * * * *", "5-1 * * * *", "1,,2 * * * *", "1/2/3 * * * *", "@daily", "? * * * *", "*/x * * * *"} {
		if _, err := ParseCron(expr); err == nil {
			t.Errorf("accepted invalid cron %q", expr)
		}
	}
	now := time.Date(2026, 3, 2, 10, 15, 0, 0, time.UTC) // Monday
	for _, expr := range []string{"*/5 * * * *", "0,10-20/5 8-18/2 * * 1-5", "15 10 * MAR MON", "15 10 1 * MON"} {
		c, err := ParseCron(expr)
		if err != nil {
			t.Fatal(expr, err)
		}
		if !c.Matches(now) {
			t.Errorf("%s does not match %v", expr, now)
		}
	}
	c, _ := ParseCron("0,15,30,45 * * * *")
	next := c.Next(now.Add(30 * time.Second))
	want := now.Add(15 * time.Minute)
	if next == nil || !next.Equal(want) || next.Location() != time.UTC {
		t.Fatalf("next = %v, want %v", next, want)
	}
	if !c.Matches(now.In(time.FixedZone("other", 3*3600))) {
		t.Fatal("cron did not use UTC")
	}
	leap, _ := ParseCron("0 0 29 2 *")
	next = leap.Next(time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))
	if next == nil || next.Year() != 2028 || next.Month() != 2 || next.Day() != 29 {
		t.Fatal("leap day", next)
	}
	impossible, _ := ParseCron("0 0 31 2 *")
	if impossible.Next(now) != nil {
		t.Fatal("impossible date has occurrence")
	}
	sunday, _ := ParseCron("0 0 * * 7")
	if !sunday.Matches(time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("Sunday 7 not supported")
	}
}

func TestScheduleClaimStatusPersistenceAndMinuteDeduplication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	s := &ServerSchedule{ID: "test", ServerID: "server", Name: "test", CronExpr: "* * * * *", Action: "backup", IsActive: true}
	if err := db.CreateServerSchedule(s); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	var wg sync.WaitGroup
	results := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := db.ClaimScheduleRun(s.ID, now, false)
			if err != nil {
				t.Error(err)
			}
			results <- ok
		}()
	}
	wg.Wait()
	close(results)
	claimed := 0
	for ok := range results {
		if ok {
			claimed++
		}
	}
	if claimed != 1 {
		t.Fatalf("%d simultaneous claims succeeded", claimed)
	}
	got, _ := db.GetServerSchedule("server", "test")
	if got.LastStatus != "started" || got.LastRunAt == nil || got.LastFinishedAt != nil {
		t.Fatalf("bad acceptance: %+v", got)
	}
	db.SetScheduleRunStatus(s.ID, "running", "")
	if ok, err := db.ClaimScheduleRun(s.ID, now.Add(time.Minute), true); err != nil || ok {
		t.Fatal("running schedule overlapped", ok, err)
	}
	db.SetScheduleRunStatus(s.ID, "failed", "backup read failed")
	got, _ = db.GetServerSchedule("server", "test")
	if got.LastStatus != "failed" || got.LastError != "backup read failed" || got.LastFinishedAt == nil {
		t.Fatal(got)
	}
	if ok, err := db.ClaimScheduleRun(s.ID, now, true); err != nil || ok {
		t.Fatal("schedule repeated in same minute", ok, err)
	}
	db.Close()
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	got, _ = db.GetServerSchedule("server", "test")
	if got.LastStatus != "failed" || got.LastError != "backup read failed" {
		t.Fatal("completion not persisted", got)
	}
	if ok, err := db.ClaimScheduleRun(s.ID, now.Add(time.Minute), true); err != nil || !ok {
		t.Fatal("next minute did not run", ok, err)
	}
}
