package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/Pekly/NeoPanel/internal/runner"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev/reverse proxy flexibility
	},
}

type ConsoleHandler struct {
	auth    *AuthHandler
	manager *runner.ServerManager
}

func NewConsoleHandler(auth *AuthHandler, manager *runner.ServerManager) *ConsoleHandler {
	return &ConsoleHandler{
		auth:    auth,
		manager: manager,
	}
}

type WSMessage struct {
	Event string `json:"event"` // "log", "command", "stats"
	Data  string `json:"data"`
}

func (h *ConsoleHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate via token query param or cookie
	token := r.URL.Query().Get("token")
	if token == "" {
		if cookie, err := r.Cookie("neopanel_token"); err == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	_, err := h.auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	serverID := getURLParam(r, "id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Send recent logs first in a structured history batch
	history := h.manager.GetLogs(serverID)
	if historyData, err := json.Marshal(history); err == nil {
		msg, _ := json.Marshal(WSMessage{Event: "history", Data: string(historyData)})
		_ = conn.WriteMessage(websocket.TextMessage, msg)
	}

	// Subscribe to live log stream
	logCh, unsubscribe, err := h.manager.Subscribe(serverID)
	if err != nil {
		_ = conn.WriteJSON(WSMessage{Event: "log", Data: "\x1b[31m[NeoPanel] Failed to subscribe to console: " + err.Error() + "\x1b[0m"})
		return
	}
	defer unsubscribe()

	// Channel to signal close
	closeCh := make(chan struct{})

	// Goroutine to send logs from logCh to WebSocket
	go func() {
		defer close(closeCh)
		for {
			select {
			case line, ok := <-logCh:
				if !ok {
					return
				}
				msg, _ := json.Marshal(WSMessage{Event: "log", Data: line})
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		}
	}()

	// Read commands from WebSocket client
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming WSMessage
		if err := json.Unmarshal(msgBytes, &incoming); err == nil && incoming.Event == "command" {
			_ = h.manager.SendCommand(serverID, incoming.Data)
		}
	}
}

