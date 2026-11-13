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
	Timezone string     `json:"timezone"`
	LastRun  *time.Time `json:"last_run"` // alias; last_run_at is retained for existing clients
}

func viewSchedule(s *database.ServerSchedule) scheduleView {
	copy := *s
	copy.NextRunAt = nil
	if copy.IsActive {
		if cron, err := database.ParseCron(copy.CronExpr); err == nil {
			copy.NextRunAt = cron.Next(time.Now().UTC())
		}
	}
	return scheduleView{&copy, "UTC", copy.LastRunAt}
}

func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.db.ListServerSchedules(getURLParam(r, "id"))
	if err != nil {
		writeError(w, 500, "Failed to list schedules: "+err.Error())
		return
	}
	views := make([]scheduleView, 0, len(schedules))
	for _, s := range schedules {
		views = append(views, viewSchedule(s))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"schedules": views})
}

type scheduleUpdate struct {
	Name     *string `json:"name"`
	CronExpr *string `json:"cron_expr"`
	Action   *string `json:"action"`
	Payload  *string `json:"payload"`
	IsActive *bool   `json:"is_active"`
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
	if s.Name == "" {
		return fmt.Errorf("schedule name is required")
	}
	if _, err := database.ParseCron(s.CronExpr); err != nil {
		return err
	}
	switch s.Action {
	case "command":
		if s.Payload == "" {
			return fmt.Errorf("command payload is required")
		}
	case "restart", "backup", "stop":
	default:
		return fmt.Errorf("supported actions: command, restart, backup, stop")
	}
	return nil
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req scheduleUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	s := &database.ServerSchedule{ID: "sch-" + randomPasswordHex(12), ServerID: getURLParam(r, "id"), IsActive: true}
	if err := req.apply(s); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if err := h.db.CreateServerSchedule(s); err != nil {
		writeError(w, 500, "Failed to create schedule: "+err.Error())
		return
	}
	logAudit(h.db, r, "schedule.create", "Created schedule "+s.Name+" on "+s.ServerID)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"status": "success", "schedule": viewSchedule(s)})
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	if err := req.apply(s); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if err := h.db.UpdateServerSchedule(s); err != nil {
		writeError(w, 500, "Failed to update schedule: "+err.Error())
		return
	}
	logAudit(h.db, r, "schedule.update", "Updated schedule "+s.ID)
	writeJSON(w, 200, map[string]interface{}{"status": "success", "schedule": viewSchedule(s)})
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
	writeJSON(w, 200, map[string]interface{}{"status": "success", "is_active": s.IsActive, "schedule": viewSchedule(s)})
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
