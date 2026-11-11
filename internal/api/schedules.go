package api

import (
	"encoding/json"
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

// ListSchedules returns all schedules for a server
func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	serverID := extractIDFromPath(r.URL.Path, "/api/servers/", "/schedules")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "Missing server ID")
		return
	}

	schedules, err := h.db.ListServerSchedules(serverID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list schedules: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"schedules": schedules,
	})
}

// CreateSchedule creates a new scheduled task
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	serverID := extractIDFromPath(r.URL.Path, "/api/servers/", "/schedules")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "Missing server ID")
		return
	}

	var req struct {
		Name     string `json:"name"`
		CronExpr string `json:"cron_expr"`
		Action   string `json:"action"`
		Payload  string `json:"payload"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "Schedule name is required")
		return
	}
	if strings.TrimSpace(req.CronExpr) == "" {
		writeError(w, http.StatusBadRequest, "Cron expression is required")
		return
	}

	// Validate action
	act := strings.ToLower(strings.TrimSpace(req.Action))
	switch act {
	case "command", "restart", "start", "stop", "backup":
		// valid
	default:
		writeError(w, http.StatusBadRequest, "Invalid action. Supported: command, restart, start, stop, backup")
		return
	}

	sched := &database.ServerSchedule{
		ID:       fmt.Sprintf("sch-%x", time.Now().UnixNano()),
		ServerID: serverID,
		Name:     strings.TrimSpace(req.Name),
		CronExpr: strings.TrimSpace(req.CronExpr),
		Action:   act,
		Payload:  strings.TrimSpace(req.Payload),
		IsActive: req.IsActive,
	}
	if err := h.db.CreateServerSchedule(sched); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create schedule: "+err.Error())
		return
	}

	_ = h.db.CreateAuditLog("admin", "schedule.create", "Created schedule '"+sched.Name+"' on server "+serverID, getClientIP(r))

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":   "success",
		"schedule": sched,
	})
}

// UpdateSchedule updates an existing schedule
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/schedules/")
	if len(parts) != 2 || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "Missing schedule ID")
		return
	}
	schedID := parts[1]

	var req struct {
		Name     string `json:"name"`
		CronExpr string `json:"cron_expr"`
		Action   string `json:"action"`
		Payload  string `json:"payload"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sched, err := h.db.GetServerScheduleByID(schedID)
	if err != nil || sched == nil {
		writeError(w, http.StatusNotFound, "Schedule not found")
		return
	}

	if strings.TrimSpace(req.Name) != "" {
		sched.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.CronExpr) != "" {
		sched.CronExpr = strings.TrimSpace(req.CronExpr)
	}
	if strings.TrimSpace(req.Action) != "" {
		act := strings.ToLower(strings.TrimSpace(req.Action))
		switch act {
		case "command", "restart", "start", "stop", "backup":
			sched.Action = act
		default:
			writeError(w, http.StatusBadRequest, "Invalid action")
			return
		}
	}
	sched.Payload = strings.TrimSpace(req.Payload)
	sched.IsActive = req.IsActive

	if err := h.db.UpdateServerSchedule(sched); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update schedule: "+err.Error())
		return
	}

	_ = h.db.CreateAuditLog("admin", "schedule.update", "Updated schedule '"+sched.Name+"' ("+schedID+")", getClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"schedule": sched,
	})
}

// DeleteSchedule deletes a schedule
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/schedules/")
	if len(parts) != 2 || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "Missing schedule ID")
		return
	}
	schedID := parts[1]

	if err := h.db.DeleteServerScheduleByID(schedID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete schedule: "+err.Error())
		return
	}

	_ = h.db.CreateAuditLog("admin", "schedule.delete", "Deleted schedule "+schedID, getClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}

// ToggleSchedule flips active state
func (h *ScheduleHandler) ToggleSchedule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/schedules/")
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "Invalid schedule URL")
		return
	}
	subparts := strings.Split(parts[1], "/toggle")
	schedID := subparts[0]

	var req struct {
		IsActive bool `json:"is_active"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.db.ToggleServerScheduleByID(schedID, req.IsActive); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to toggle schedule: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "success",
		"is_active": req.IsActive,
	})
}

// RunScheduleNow triggers immediate execution of the schedule
func (h *ScheduleHandler) RunScheduleNow(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/schedules/")
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "Invalid schedule URL")
		return
	}
	subparts := strings.Split(parts[1], "/run")
	schedID := subparts[0]

	sched, err := h.db.GetServerScheduleByID(schedID)
	if err != nil || sched == nil {
		writeError(w, http.StatusNotFound, "Schedule not found")
		return
	}

	_ = h.db.RecordScheduleRun(sched.ID, time.Now().UTC(), nil)
	go h.manager.ExecuteSchedule(sched)

	_ = h.db.CreateAuditLog("admin", "schedule.run_now", "Manually executed schedule '"+sched.Name+"'", getClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Schedule execution triggered successfully",
	})
}

