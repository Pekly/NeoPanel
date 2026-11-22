package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

// PowerCountdownState is deliberately transient: restarting the panel cancels it.
type PowerCountdownState struct {
	Action    string    `json:"action"`
	ExecuteAt time.Time `json:"execute_at"`
	Message   string    `json:"message"`
}

type countdownPowerManager interface {
	GetPowerRun(string) (runner.PowerRun, bool)
	SendCommand(string, string) error
	CountdownPower(*database.Server, string, runner.PowerRun) error
}

type countdownTimer interface{ Stop() bool }

type pendingPowerCountdown struct {
	PowerCountdownState
	origin      JWTClaims
	run         runner.PowerRun
	timer       countdownTimer
	lastSeconds int
}

// All countdown mutations, direct HTTP power actions and deletion share mu.
// Hold it through the final manager call: cancellation either wins before that
// call or observes its completion, never races an already-detached action.
// The manager still owns its existing restore/power operation gates.
type powerCountdownCoordinator struct {
	mu        sync.Mutex
	db        *database.DB
	manager   countdownPowerManager
	pending   map[string]*pendingPowerCountdown
	now       func() time.Time
	afterFunc func(time.Duration, func()) countdownTimer
}

func (h *ServerHandler) powerCountdowns() *powerCountdownCoordinator {
	h.countdownOnce.Do(func() {
		h.countdowns = &powerCountdownCoordinator{
			db: h.db, manager: h.manager,
			pending: make(map[string]*pendingPowerCountdown), now: time.Now,
			afterFunc: func(d time.Duration, f func()) countdownTimer { return time.AfterFunc(d, f) },
		}
	})
	return h.countdowns
}

func (c *powerCountdownCoordinator) cancelLocked(id string) {
	if p := c.pending[id]; p != nil {
		if p.timer != nil {
			p.timer.Stop()
		}
		delete(c.pending, id)
	}
}

// Re-read the user and grants, not the original session's cached role. A user
// promoted since scheduling cannot acquire admin authority for this job.
// ID zero is the existing authenticated node-daemon (not a human account).
func (c *powerCountdownCoordinator) authorized(id string, p *pendingPowerCountdown) bool {
	if !c.db.FeatureEnabled(id, "countdowns") {
		return false
	}
	fresh := p.origin
	if fresh.UserID > 0 {
		user, err := c.db.GetUserByID(fresh.UserID)
		if err != nil || user == nil {
			return false
		}
		if user.Role != "admin" || fresh.Role != "admin" {
			fresh.Role = "user"
		}
	} else if fresh.UserID != 0 || fresh.Role != "admin" {
		return false
	}
	return authorizeServer(c.db, &fresh, id, "server.power+console.command") == 0
}

func (c *powerCountdownCoordinator) expireLocked(id string) {
	if p := c.pending[id]; p != nil {
		run, online := c.manager.GetPowerRun(id)
		if !online || run != p.run || !c.authorized(id, p) {
			c.cancelLocked(id)
		}
	}
}

func countdownWarning(p *pendingPowerCountdown, seconds int) string {
	text := fmt.Sprintf("Server will %s in %d seconds.", p.Action, seconds)
	if p.Message != "" {
		text += " " + p.Message
	}
	return "say " + text
}

// A single one-shot timer per pending server checks liveness/authorization once
// a second. No permanently running ticker/goroutine or request context is kept.
func (c *powerCountdownCoordinator) armLocked(id string, p *pendingPowerCountdown) {
	delay := p.ExecuteAt.Sub(c.now())
	if delay > time.Second {
		delay = time.Second
	}
	if delay < 0 {
		delay = 0
	}
	p.timer = c.afterFunc(delay, func() { c.tick(id, p) })
}

func (c *powerCountdownCoordinator) tick(id string, p *pendingPowerCountdown) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// A stopped timer may already have queued its callback. Pointer identity
	// prevents that callback from affecting a replacement countdown.
	if c.pending[id] != p {
		return
	}
	c.expireLocked(id)
	if c.pending[id] != p {
		return
	}
	remaining := int(math.Ceil(p.ExecuteAt.Sub(c.now()).Seconds()))
	if remaining <= 0 {
		defer c.cancelLocked(id)
		srv, err := c.db.GetServer(id)
		if err == nil {
			err = c.manager.CountdownPower(srv, p.Action, p.run)
		}
		if err != nil {
			log.Printf("Power countdown %s on %s failed: %v", p.Action, id, err)
		}
		return
	}
	// Send at most one warning after a delayed wakeup, never stale milestones.
	for _, milestone := range []int{10, 30, 60} {
		if remaining <= milestone && p.lastSeconds > milestone {
			if err := c.manager.SendCommand(id, countdownWarning(p, remaining)); err != nil {
				c.cancelLocked(id)
				return
			}
			break
		}
	}
	p.lastSeconds = remaining
	c.armLocked(id, p)
}

func (h *ServerHandler) PowerCountdown(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if status := authorizeServer(h.db, claims, id, serverActionPermission(r.Method, "power/countdown")); status != 0 {
		writeError(w, status, "Access denied to power countdown")
		return
	}
	var req struct {
		Action  string          `json:"action"`
		Seconds int             `json:"seconds"`
		Message json.RawMessage `json:"message"`
	}
	message := ""
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid countdown request")
			return
		}
		if err := decoder.Decode(new(interface{})); err != io.EOF {
			writeError(w, http.StatusBadRequest, "Expected a single countdown request")
			return
		}
		if len(req.Message) != 0 {
			if string(req.Message) == "null" || json.Unmarshal(req.Message, &message) != nil {
				writeError(w, http.StatusBadRequest, "Message must be a string")
				return
			}
		}
		if (req.Action != "restart" && req.Action != "stop") || req.Seconds < 10 || req.Seconds > 3600 {
			writeError(w, http.StatusBadRequest, "Action must be restart or stop; seconds must be an integer from 10 to 3600")
			return
		}
		if utf8.RuneCountInString(message) > 200 || strings.ContainsFunc(message, unicode.IsControl) {
			writeError(w, http.StatusBadRequest, "Message must be at most 200 characters without control characters")
			return
		}
	}

	c := h.powerCountdowns()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expireLocked(id)
	switch r.Method {
	case http.MethodPost:
		if c.pending[id] != nil {
			writeError(w, http.StatusConflict, "A power countdown is already pending; cancel it first")
			return
		}
		run, online := c.manager.GetPowerRun(id)
		if !online {
			writeError(w, http.StatusConflict, "Server must be online to warn players")
			return
		}
		p := &pendingPowerCountdown{
			PowerCountdownState: PowerCountdownState{Action: req.Action, ExecuteAt: c.now().Add(time.Duration(req.Seconds) * time.Second).UTC(), Message: message},
			origin:              *claims, run: run, lastSeconds: req.Seconds,
		}
		if !c.authorized(id, p) {
			writeError(w, http.StatusForbidden, "Countdown permissions are no longer available")
			return
		}
		if err := c.manager.SendCommand(id, countdownWarning(p, req.Seconds)); err != nil {
			writeError(w, http.StatusConflict, "Could not warn players: "+err.Error())
			return
		}
		c.pending[id] = p
		c.armLocked(id, p)
		logAudit(h.db, r, "power_countdown_create", fmt.Sprintf("Scheduled %s on %s in %d seconds", req.Action, id, req.Seconds))
	case http.MethodDelete:
		c.cancelLocked(id)
		logAudit(h.db, r, "power_countdown_cancel", "Cancelled power countdown on "+id)
	case http.MethodGet:
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var state *PowerCountdownState
	if p := c.pending[id]; p != nil {
		copy := p.PowerCountdownState
		state = &copy
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, struct {
		Countdown *PowerCountdownState `json:"countdown"`
	}{state})
}
