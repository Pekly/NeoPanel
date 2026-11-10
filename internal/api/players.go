package api

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	Name     string `json:"name"`
	UUID     string `json:"uuid,omitempty"`
	IP       string `json:"ip,omitempty"`
	Playtime string `json:"playtime,omitempty"`
	Ping     int    `json:"ping,omitempty"`
	OP       bool   `json:"op,omitempty"`
}

type BanInfo struct {
	Name    string `json:"name"`
	Reason  string `json:"reason"`
	Created string `json:"created,omitempty"`
}

type PlayersOverview struct {
	Online       []PlayerInfo `json:"online"`
	Ops          []PlayerInfo `json:"ops"`
	Whitelist    []PlayerInfo `json:"whitelist"`
	Bans         []BanInfo    `json:"bans"`
	OnlineCount  int          `json:"online_count"`
	MaxPlayers   int          `json:"max_players"`
	ServerStatus string       `json:"server_status"`
}

func formatSessionDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("%dm", mins)
	}
	hrs := mins / 60
	remMins := mins % 60
	return fmt.Sprintf("%dh %dm", hrs, remMins)
}

func writeMCVarInt(buf *bytes.Buffer, v int) {
	u := uint32(v)
	for {
		if (u & ^uint32(0x7F)) == 0 {
			buf.WriteByte(byte(u))
			return
		}
		buf.WriteByte(byte((u & 0x7F) | 0x80))
		u >>= 7
	}
}

func readMCVarInt(r io.Reader) (int, error) {
	var num int
	var shift uint
	var b [1]byte
	for i := 0; i < 5; i++ {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}
		num |= int(b[0]&0x7F) << shift
		if (b[0] & 0x80) == 0 {
			return num, nil
		}
		shift += 7
	}
	return 0, fmt.Errorf("varint too big")
}

type slpPingResult struct {
	Online int
	Max    int
	Sample []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	LatencyMS int
}

// queryLocalServerSLP performs a lightweight Minecraft Server List Ping (1.7+) to localhost:<port>
func queryLocalServerSLP(port int) (*slpPingResult, error) {
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port")
	}

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 350*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(450 * time.Millisecond))

	var handshake bytes.Buffer
	handshake.WriteByte(0x00)      // Packet ID 0x00 (Handshake)
	writeMCVarInt(&handshake, 763) // Protocol version (1.20.1 / generic SLP)
	hostStr := "127.0.0.1"
	writeMCVarInt(&handshake, len(hostStr))
	handshake.WriteString(hostStr)
	_ = binary.Write(&handshake, binary.BigEndian, uint16(port))
	writeMCVarInt(&handshake, 1) // Next state: 1 (Status)

	var frame bytes.Buffer
	writeMCVarInt(&frame, handshake.Len())
	frame.Write(handshake.Bytes())
	// Status Request (length 1, packet ID 0x00)
	frame.WriteByte(0x01)
	frame.WriteByte(0x00)

	started := time.Now()
	if _, err := conn.Write(frame.Bytes()); err != nil {
		return nil, err
	}

	// Packet length & ID
	if _, err := readMCVarInt(conn); err != nil {
		return nil, err
	}
	pktID, err := readMCVarInt(conn)
	if err != nil || pktID != 0x00 {
		return nil, fmt.Errorf("unexpected SLP packet ID")
	}
	jsonLen, err := readMCVarInt(conn)
	if err != nil || jsonLen <= 0 || jsonLen > 256*1024 {
		return nil, fmt.Errorf("invalid SLP JSON length")
	}
	rawJSON := make([]byte, jsonLen)
	if _, err := io.ReadFull(conn, rawJSON); err != nil {
		return nil, err
	}
	latency := int(time.Since(started).Milliseconds())
	if latency < 1 {
		latency = 1
	}

	var parsed struct {
		Players struct {
			Max    int `json:"max"`
			Online int `json:"online"`
			Sample []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"sample"`
		} `json:"players"`
	}
	if err := json.Unmarshal(rawJSON, &parsed); err != nil {
		return nil, err
	}

	return &slpPingResult{
		Online:    parsed.Players.Online,
		Max:       parsed.Players.Max,
		Sample:    parsed.Players.Sample,
		LatencyMS: latency,
	}, nil
}

func (h *PlayerHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dir := h.manager.GetWorkingDir(serverID)
	srv, _ := h.db.GetServer(serverID)
	status := h.manager.GetStatus(serverID)

	overview := PlayersOverview{
		Online:       []PlayerInfo{},
		Ops:          []PlayerInfo{},
		Whitelist:    []PlayerInfo{},
		Bans:         []BanInfo{},
		ServerStatus: status,
		MaxPlayers:   20,
	}

	opSet := make(map[string]bool)
	opUUID := make(map[string]string)

	// 1. Read ops.json
	if data, err := os.ReadFile(filepath.Join(dir, "ops.json")); err == nil {
		var rawOps []struct {
			Name string `json:"name"`
			UUID string `json:"uuid"`
		}
		if json.Unmarshal(data, &rawOps) == nil {
			for _, op := range rawOps {
				overview.Ops = append(overview.Ops, PlayerInfo{Name: op.Name, UUID: op.UUID, OP: true})
				opSet[strings.ToLower(op.Name)] = true
				if op.UUID != "" {
					opUUID[strings.ToLower(op.Name)] = op.UUID
				}
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

	// 4. Populate real online players from runner console parser + localhost SLP fallback
	var slp *slpPingResult
	if (status == "online" || status == "starting") && srv != nil && srv.Port > 0 {
		if pingRes, err := queryLocalServerSLP(srv.Port); err == nil {
			slp = pingRes
			if pingRes.Max > 0 {
				overview.MaxPlayers = pingRes.Max
			}
		}
	}

	if status == "online" || status == "starting" {
		tracked := h.manager.GetOnlinePlayers(serverID)
		if len(tracked) > 0 {
			for _, tp := range tracked {
				low := strings.ToLower(tp.Name)
				uuid := tp.UUID
				if uuid == "" {
					uuid = opUUID[low]
				}
				pt := ""
				if !tp.JoinedAt.IsZero() {
					pt = formatSessionDuration(time.Since(tp.JoinedAt))
				}
				pingVal := 0
				if slp != nil {
					pingVal = slp.LatencyMS
				}
				overview.Online = append(overview.Online, PlayerInfo{
					Name:     tp.Name,
					UUID:     uuid,
					IP:       tp.IP,
					Playtime: pt,
					Ping:     pingVal,
					OP:       opSet[low],
				})
			}
		} else if slp != nil && len(slp.Sample) > 0 {
			for _, sp := range slp.Sample {
				if strings.TrimSpace(sp.Name) == "" {
					continue
				}
				low := strings.ToLower(sp.Name)
				uuid := sp.ID
				if uuid == "" {
					uuid = opUUID[low]
				}
				overview.Online = append(overview.Online, PlayerInfo{
					Name:     sp.Name,
					UUID:     uuid,
					Playtime: "Active",
					Ping:     slp.LatencyMS,
					OP:       opSet[low],
				})
			}
		}
	}

	overview.OnlineCount = len(overview.Online)
	if overview.OnlineCount == 0 && slp != nil && slp.Online > 0 {
		overview.OnlineCount = slp.Online
	}

	writeJSON(w, http.StatusOK, overview)
}

func (h *PlayerHandler) Action(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")

	var req struct {
		Action string `json:"action"` // op, deop, kick, ban, pardon, whitelist_add, whitelist_remove, sync
		Player string `json:"player"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Action == "sync" || req.Action == "refresh" {
		if err := h.manager.SendCommand(serverID, "list"); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to sync players: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Synced player list via console"})
		return
	}

	if strings.TrimSpace(req.Player) == "" {
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
