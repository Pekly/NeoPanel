package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type PlayerHandler struct {
	db      *database.DB
	manager *runner.ServerManager
}

func NewPlayerHandler(db *database.DB, manager *runner.ServerManager) *PlayerHandler {
	return &PlayerHandler{
		db:      db,
		manager: manager,
	}
}

type PlayerInfo struct {
	Name string `json:"name"`
	UUID string `json:"uuid,omitempty"`
}

type BanInfo struct {
	Name    string `json:"name"`
	Reason  string `json:"reason"`
	Created string `json:"created,omitempty"`
}

type PlayersOverview struct {
	Online    []PlayerInfo `json:"online"`
	Ops       []PlayerInfo `json:"ops"`
	Whitelist []PlayerInfo `json:"whitelist"`
	Bans      []BanInfo    `json:"bans"`
}

func (h *PlayerHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dir := h.manager.GetWorkingDir(serverID)

	overview := PlayersOverview{
		Online:    []PlayerInfo{},
		Ops:       []PlayerInfo{},
		Whitelist: []PlayerInfo{},
		Bans:      []BanInfo{},
	}

	// 1. Read ops.json
	if data, err := os.ReadFile(filepath.Join(dir, "ops.json")); err == nil {
		var rawOps []struct {
			Name string `json:"name"`
			UUID string `json:"uuid"`
		}
		if json.Unmarshal(data, &rawOps) == nil {
			for _, op := range rawOps {
				overview.Ops = append(overview.Ops, PlayerInfo{Name: op.Name, UUID: op.UUID})
			}
		}
	}

	// 2. Read whitelist.json
	if data, err := os.ReadFile(filepath.Join(dir, "whitelist.json")); err == nil {
		var rawWl []struct {
			Name string `json:"name"`
			UUID string `json:"uuid"`
		}
		if json.Unmarshal(data, &rawWl) == nil {
			for _, p := range rawWl {
				overview.Whitelist = append(overview.Whitelist, PlayerInfo{Name: p.Name, UUID: p.UUID})
			}
		}
	}

	// 3. Read banned-players.json
	if data, err := os.ReadFile(filepath.Join(dir, "banned-players.json")); err == nil {
		var rawBans []struct {
			Name    string `json:"name"`
			Reason  string `json:"reason"`
			Created string `json:"created"`
		}
		if json.Unmarshal(data, &rawBans) == nil {
			for _, b := range rawBans {
				overview.Bans = append(overview.Bans, BanInfo{Name: b.Name, Reason: b.Reason, Created: b.Created})
			}
		}
	}

	writeJSON(w, http.StatusOK, overview)
}

func (h *PlayerHandler) Action(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")

	var req struct {
		Action string `json:"action"` // op, deop, kick, ban, pardon, whitelist_add, whitelist_remove
		Player string `json:"player"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Player) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid player name and action are required"})
		return
	}

	player := strings.TrimSpace(req.Player)
	var cmd string

	switch req.Action {
	case "op":
		cmd = "op " + player
	case "deop":
		cmd = "deop " + player
	case "kick":
		reason := req.Reason
		if reason == "" {
			reason = "Kicked by administrator"
		}
		cmd = "kick " + player + " " + reason
	case "ban":
		reason := req.Reason
		if reason == "" {
			reason = "Banned by administrator"
		}
		cmd = "ban " + player + " " + reason
	case "pardon":
		cmd = "pardon " + player
	case "whitelist_add":
		cmd = "whitelist add " + player
	case "whitelist_remove":
		cmd = "whitelist remove " + player
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown action: " + req.Action})
		return
	}

	if err := h.manager.SendCommand(serverID, cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to dispatch command: " + err.Error()})
		return
	}

	logAudit(h.db, r, "player_"+req.Action, fmt.Sprintf("Executed '%s' on %s", cmd, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Player action executed: " + cmd,
	})
}

