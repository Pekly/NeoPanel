package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 16384,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		return err == nil && strings.EqualFold(u.Host, r.Host)
	},
}

type ConsoleHandler struct {
	db      *database.DB
	auth    *AuthHandler
	manager *runner.ServerManager
}

func NewConsoleHandler(db *database.DB, auth *AuthHandler, manager *runner.ServerManager) *ConsoleHandler {
	return &ConsoleHandler{
		db:      db,
		auth:    auth,
		manager: manager,
	}
}

type WSMessage struct {
	Event string `json:"event"` // "log", "command", "stats", "ping", "pong"
	Data  string `json:"data"`
}

type WSStatsPayload struct {
	Event           string  `json:"event"`
	ServerID        string  `json:"server_id"`
	Status          string  `json:"status"`
	CPUPercent      float64 `json:"cpu_percent"`
	MemoryMB        int64   `json:"memory_mb"`
	MemoryMaxMB     int     `json:"memory_max_mb"`
	DiskMB          int64   `json:"disk_mb"`
	UptimeSec       int64   `json:"uptime_sec"`
	NetInKiB        float64 `json:"net_in_kib"`
	NetOutKiB       float64 `json:"net_out_kib"`
	StartupFailures int     `json:"startup_failures"`
	StartupHalted   bool    `json:"startup_halted"`
}

func (h *ConsoleHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate via token query param or cookie
	token := r.URL.Query().Get("token")
	if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {
		token = strings.TrimPrefix(bearer, "Bearer ")
	}
	if token == "" {
		if cookie, err := r.Cookie("neopanel_token"); err == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	claims, err := h.auth.ValidateAuthToken(token)
	if err != nil {
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}
	claims, err = forwardedClaims(h.db, r, claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	serverID := getURLParam(r, "id")
	if status := authorizeServer(h.db, claims, serverID, "console.view"); status != 0 {
		http.Error(w, "Console access denied", status)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	const (
		writeWait  = 10 * time.Second
		pongWait   = 65 * time.Second
		pingPeriod = 20 * time.Second
	)

	conn.SetReadLimit(65536)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Subscribe to live log stream FIRST so no lines between GetLogs and Subscribe are missed
	logCh, unsubscribe, err := h.manager.Subscribe(serverID)
	if err != nil {
		_ = conn.WriteJSON(WSMessage{Event: "log", Data: "\x1b[31m[NeoPanel] Failed to subscribe to console: " + err.Error() + "\x1b[0m"})
		return
	}
	defer unsubscribe()

	// Send initial history logs + immediate initial stats snapshot before starting writer loop
	history := h.manager.GetLogs(serverID)
	if historyData, err := json.Marshal(history); err == nil {
		msg, _ := json.Marshal(WSMessage{Event: "history", Data: string(historyData)})
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}

	if st, ok := h.manager.GetStatsByID(serverID); ok {
		payload := WSStatsPayload{
			Event:           "stats",
			ServerID:        st.ServerID,
			Status:          st.Status,
			CPUPercent:      st.CPUUsage,
			MemoryMB:        st.MemoryMB,
			MemoryMaxMB:     st.MemoryMaxMB,
			DiskMB:          st.DiskMB,
			UptimeSec:       st.UptimeSec,
			NetInKiB:        st.NetInKiB,
			NetOutKiB:       st.NetOutKiB,
			StartupFailures: st.StartupFailures,
			StartupHalted:   st.StartupHalted,
		}
		if b, err := json.Marshal(payload); err == nil {
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return
			}
		}
	}

	doneCh := make(chan struct{})
	pongReplyCh := make(chan struct{}, 4)

	// Single dedicated writer goroutine (guarantees gorilla/websocket concurrency safety)
	go func() {
		statsTicker := time.NewTicker(1 * time.Second)
		pingTicker := time.NewTicker(pingPeriod)
		defer statsTicker.Stop()
		defer pingTicker.Stop()

		for {
			select {
			case <-doneCh:
				return

			case line, ok := <-logCh:
				if !ok {
					_ = conn.Close()
					return
				}
				msg, _ := json.Marshal(WSMessage{Event: "log", Data: line})
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					_ = conn.Close()
					return
				}

			case <-statsTicker.C:
				if authorizeServer(h.db, claims, serverID, "console.view") != 0 {
					_ = conn.Close()
					return
				}
				st, ok := h.manager.GetStatsByID(serverID)
				if !ok {
					continue
				}
				payload := WSStatsPayload{
					Event:           "stats",
					ServerID:        st.ServerID,
					Status:          st.Status,
					CPUPercent:      st.CPUUsage,
					MemoryMB:        st.MemoryMB,
					MemoryMaxMB:     st.MemoryMaxMB,
					DiskMB:          st.DiskMB,
					UptimeSec:       st.UptimeSec,
					NetInKiB:        st.NetInKiB,
					NetOutKiB:       st.NetOutKiB,
					StartupFailures: st.StartupFailures,
					StartupHalted:   st.StartupHalted,
				}
				b, _ := json.Marshal(payload)
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					_ = conn.Close()
					return
				}

			case <-pongReplyCh:
				b, _ := json.Marshal(WSMessage{Event: "pong"})
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					_ = conn.Close()
					return
				}

			case <-pingTicker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					_ = conn.Close()
					return
				}
			}
		}
	}()

	// Reader loop: handles incoming console commands and client heartbeats
	defer close(doneCh)
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))

		var incoming WSMessage
		if err := json.Unmarshal(msgBytes, &incoming); err == nil {
			switch incoming.Event {
			case "command":
				// Re-check each command so permission revocation takes effect on open sockets.
				if authorizeServer(h.db, claims, serverID, "console.command") != 0 {
					return
				}
				cmd := strings.TrimSpace(incoming.Data)
				if cmd != "" {
					if err := h.manager.SendCommand(serverID, incoming.Data); err == nil {
						uid := 0
						if claims != nil {
							uid = claims.UserID
						}
						logAuditUser(h.db, uid, r, "console_command", fmt.Sprintf("Executed command on %s: %s", auditServerLabel(h.db, serverID), truncateAuditText(cmd, 120)))
					}
				}
			case "ping":
				select {
				case pongReplyCh <- struct{}{}:
				default:
				}
			}
		}
	}
}
