package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type ScheduleHandler struct {
	db      *database.DB
	manager *runner.ServerManager
}

func NewScheduleHandler(db *database.DB, manager *runner.ServerManager) *ScheduleHandler {
	return &ScheduleHandler{db: db, manager: manager}
}

type scheduleView struct {
	*database.ServerSchedule
	Timezone  string                  `json:"timezone"`
	LastRun   *time.Time              `json:"last_run"` // alias; last_run_at is retained for existing clients
	Tasks     []database.ScheduleTask `json:"tasks"`
	TaskCount int                     `json:"task_count"`
}

func viewSchedule(s *database.ServerSchedule) scheduleView {
	copy := *s
	copy.NextRunAt = nil
	if copy.IsActive {
		if cron, err := database.ParseCron(copy.CronExpr); err == nil {
			copy.NextRunAt = cron.Next(time.Now().UTC())
		}
	}
	return scheduleView{ServerSchedule: &copy, Timezone: "UTC", LastRun: copy.LastRunAt, Tasks: []database.ScheduleTask{{Order: 1, Action: s.Action, Payload: s.Payload}}, TaskCount: 1}
}

func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.db.ListServerSchedules(getURLParam(r, "id"))
	if err != nil {
		writeError(w, 500, "Failed to list schedules: "+err.Error())
		return
	}
	views := make([]scheduleView, 0, len(schedules))
	for _, s := range schedules {
		v, err := h.scheduleView(s)
		if err != nil {
			writeError(w, 500, "Failed to load schedule tasks: "+err.Error())
			return
		}
		views = append(views, v)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"schedules": views})
}

type scheduleUpdate struct {
	Name     *string         `json:"name"`
	CronExpr *string         `json:"cron_expr"`
	Action   *string         `json:"action"`
	Payload  *string         `json:"payload"`
	IsActive *bool           `json:"is_active"`
	Tasks    json.RawMessage `json:"tasks"`
}

func (u scheduleUpdate) apply(s *database.ServerSchedule) error {
	if u.Name != nil {
		s.Name = strings.TrimSpace(*u.Name)
	}
	if u.CronExpr != nil {
		s.CronExpr = strings.TrimSpace(*u.CronExpr)
	}
	if u.Action != nil {
		s.Action = strings.ToLower(strings.TrimSpace(*u.Action))
	}
	if u.Payload != nil {
		s.Payload = strings.TrimSpace(*u.Payload)
	}
	if u.IsActive != nil {
		s.IsActive = *u.IsActive
	}
	if u.Tasks != nil {
		tasks, err := u.parseTasks()
		if err != nil {
			return err
		}
		s.Action, s.Payload = tasks[0].Action, tasks[0].Payload
	}
	if s.Name == "" {
		return fmt.Errorf("schedule name is required")
	}
	if _, err := database.ParseCron(s.CronExpr); err != nil {
		return err
	}
	return database.ValidateScheduleTasks([]database.ScheduleTask{{Action: s.Action, Payload: s.Payload}})
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req scheduleUpdate
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10)).Decode(&req); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	s := &database.ServerSchedule{ID: "sch-" + randomPasswordHex(12), ServerID: getURLParam(r, "id"), IsActive: true}
	if err := req.apply(s); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	tasks, _ := req.parseTasks() // apply already validated the request
	if err := h.db.SaveScheduleWithTasks(s, tasks, true); err != nil {
		writeError(w, 500, "Failed to create schedule: "+err.Error())
		return
	}
	logAudit(h.db, r, "schedule.create", "Created schedule "+s.Name+" on "+s.ServerID)
	h.respondSchedule(w, http.StatusCreated, s)
}

func (h *ScheduleHandler) scopedSchedule(w http.ResponseWriter, r *http.Request) *database.ServerSchedule {
	s, err := h.db.GetServerSchedule(getURLParam(r, "id"), getURLParam(r, "schedId"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 404, "Schedule not found")
		} else {
			writeError(w, 500, "Failed to load schedule")
		}
		return nil
	}
	return s
}

func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	s := h.scopedSchedule(w, r)
	if s == nil {
		return
	}
	var req scheduleUpdate
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10)).Decode(&req); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	if err := req.apply(s); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	tasks, _ := req.parseTasks()
	if req.Tasks == nil && (req.Action != nil || req.Payload != nil) {
		var err error
		tasks, err = h.db.ListScheduleTasks(s)
		if err != nil {
			writeError(w, 500, "Failed to load schedule tasks: "+err.Error())
			return
		}
		tasks[0].Action, tasks[0].Payload = s.Action, s.Payload
	}
	if err := h.db.SaveScheduleWithTasks(s, tasks, false); err != nil {
		writeError(w, 500, "Failed to update schedule: "+err.Error())
		return
	}
	logAudit(h.db, r, "schedule.update", "Updated schedule "+s.ID)
	h.respondSchedule(w, http.StatusOK, s)
}

func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	s := h.scopedSchedule(w, r)
	if s == nil {
		return
	}
	if s.LastStatus == "started" || s.LastStatus == "running" {
		writeError(w, 409, "Cannot delete a running schedule")
		return
	}
	if err := h.db.DeleteServerSchedule(s.ServerID, s.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 409, "Schedule was removed or started running; refresh and retry")
			return
		}
		writeError(w, 500, "Failed to delete schedule: "+err.Error())
		return
	}
	logAudit(h.db, r, "schedule.delete", "Deleted schedule "+s.ID)
	writeJSON(w, 200, map[string]string{"status": "success"})
}

func (h *ScheduleHandler) ToggleSchedule(w http.ResponseWriter, r *http.Request) {
	s := h.scopedSchedule(w, r)
	if s == nil {
		return
	}
	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IsActive == nil {
		writeError(w, 400, "is_active is required")
		return
	}
	if *req.IsActive {
		if _, err := database.ParseCron(s.CronExpr); err != nil {
			writeError(w, 400, err.Error())
			return
		}
	}
	if err := h.db.ToggleServerSchedule(s.ServerID, s.ID, *req.IsActive); err != nil {
		writeError(w, 500, "Failed to toggle schedule: "+err.Error())
		return
	}
	s.IsActive = *req.IsActive
	h.respondSchedule(w, http.StatusOK, s)
}

func (h *ScheduleHandler) RunScheduleNow(w http.ResponseWriter, r *http.Request) {
	s := h.scopedSchedule(w, r)
	if s == nil {
		return
	}
	if _, err := database.ParseCron(s.CronExpr); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if err := h.manager.StartSchedule(s); err != nil {
		status := 500
		if errors.Is(err, runner.ErrScheduleRunning) {
			status = 409
		}
		writeError(w, status, "Failed to start schedule: "+err.Error())
		return
	}
	logAudit(h.db, r, "schedule.run_now", "Started schedule "+s.ID)
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "started", "message": "Schedule accepted; inspect last_status for completion"})
}

func (u scheduleUpdate) parseTasks() ([]database.ScheduleTask, error) {
	if u.Tasks == nil {
		return nil, nil
	}
	var tasks []database.ScheduleTask
	if err := json.Unmarshal(u.Tasks, &tasks); err != nil {
		return nil, fmt.Errorf("invalid tasks: %w", err)
	}
	var raw []map[string]json.RawMessage
	if err := json.Unmarshal(u.Tasks, &raw); err != nil {
		return nil, fmt.Errorf("invalid task objects: %w", err)
	}
	for i, fields := range raw {
		for _, field := range []string{"action", "payload", "time_offset", "continue_on_failure"} {
			if string(fields[field]) == "null" {
				return nil, fmt.Errorf("task %d: %s must not be null", i+1, field)
			}
		}
	}
	for i := range tasks {
		tasks[i].Action = strings.ToLower(strings.TrimSpace(tasks[i].Action))
		tasks[i].Payload = strings.TrimSpace(tasks[i].Payload)
	}
	if err := database.ValidateScheduleTasks(tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (h *ScheduleHandler) scheduleView(s *database.ServerSchedule) (scheduleView, error) {
	v := viewSchedule(s)
	tasks, err := h.db.ListScheduleTasks(s)
	v.Tasks, v.TaskCount = tasks, len(tasks)
	return v, err
}

func (h *ScheduleHandler) respondSchedule(w http.ResponseWriter, status int, s *database.ServerSchedule) {
	v, err := h.scheduleView(s)
	if err != nil {
		writeError(w, 500, "Failed to load schedule tasks: "+err.Error())
		return
	}
	writeJSON(w, status, map[string]interface{}{"status": "success", "is_active": s.IsActive, "schedule": v})
}
