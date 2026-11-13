package api

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestScheduleTaskAPIValidationAndPartialUpdates(t *testing.T) {
	f := newReliabilityFixture(t)
	if err := f.db.MigrateScheduleTasks(); err != nil {
		t.Fatal(err)
	}
	path := "/api/servers/alpha/schedules"
	invalid := []string{
		`[]`, `null`, `{}`, `[null]`, `[{"action":"unknown"}]`, `[{"action":"command"}]`,
		`[{"action":"backup","time_offset":-1}]`, `[{"action":"backup","time_offset":3601}]`,
		`[{"action":"backup","time_offset":0.5}]`, `[{"action":"backup","time_offset":"1"}]`,
		`[{"action":"backup","time_offset":null}]`, `[{"action":"backup","continue_on_failure":"true"}]`,
		`[{"action":"command","payload":"` + strings.Repeat("x", 4097) + `"}]`,
		`[` + strings.TrimSuffix(strings.Repeat(`{"action":"backup"},`, 21), ",") + `]`,
	}
	for _, tasks := range invalid {
		body := `{"name":"test","cron_expr":"* * * * *","tasks":` + tasks + `}`
		if w := f.request("POST", path, body, "owner"); w.Code != 400 {
			t.Fatalf("invalid tasks accepted %s: %d %s", tasks, w.Code, w.Body.String())
		}
	}
	body := `{"name":"multi","cron_expr":"* * * * *","tasks":[{"action":"command","payload":"say hello","time_offset":3600,"continue_on_failure":true},{"action":"start"},{"action":"stop"},{"action":"restart"},{"action":"backup"}]}`
	w := f.request("POST", path, body, "owner")
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result struct {
		Schedule scheduleView `json:"schedule"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	before := result.Schedule.Tasks
	if len(before) != 5 || before[0].ID == 0 || result.Schedule.Action != "command" {
		t.Fatal(result)
	}
	id := result.Schedule.ID
	w = f.request("PUT", path+"/"+id, `{"name":"renamed","is_active":false}`, "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	if !reflect.DeepEqual(before, result.Schedule.Tasks) || result.Schedule.IsActive {
		t.Fatal("partial update lost tasks", result)
	}
	w = f.request("PUT", path+"/"+id, `{"payload":"say changed"}`, "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result.Schedule.Tasks) != 5 || result.Schedule.Tasks[0].Payload != "say changed" || result.Schedule.Tasks[1].Action != "start" {
		t.Fatal("legacy update lost steps", result)
	}
	w = f.request("POST", path, `{"name":"legacy","cron_expr":"* * * * *","action":"backup"}`, "owner")
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result.Schedule.Tasks) != 1 || result.Schedule.Tasks[0].Action != "backup" || result.Schedule.Tasks[0].Order != 1 {
		t.Fatal(result)
	}
	w = f.request("GET", path, "", "owner")
	var list struct {
		Schedules []scheduleView `json:"schedules"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list.Schedules) != 2 {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	for _, s := range list.Schedules {
		if len(s.Tasks) == 0 || s.TaskCount != len(s.Tasks) {
			t.Fatal(s)
		}
	}
}

func TestScheduleTaskValidationBoundary(t *testing.T) {
	tasks := make([]database.ScheduleTask, 20)
	for i := range tasks {
		tasks[i] = database.ScheduleTask{Action: "command", Payload: strings.Repeat("x", 4096), TimeOffset: 3600}
	}
	if err := database.ValidateScheduleTasks(tasks); err != nil {
		t.Fatal(err)
	}
}
