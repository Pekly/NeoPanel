package api

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/gorilla/websocket"
)

func TestConsoleLocalAndProxyPermissionEnforcement(t *testing.T) {
	master := newReliabilityFixture(t)
	node := newReliabilityFixture(t)
	nodeHTTP := httptest.NewServer(node.router)
	defer nodeHTTP.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "remote", BaseURL: nodeHTTP.URL, APIToken: node.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	masterHTTP := httptest.NewServer(master.router)
	defer masterHTTP.Close()
	for _, endpoint := range []string{"/api/servers/alpha/ws", "/api/nodes/remote/ws/alpha"} {
		t.Run(endpoint, func(t *testing.T) {
			address := "ws" + strings.TrimPrefix(masterHTTP.URL, "http") + endpoint
			header := http.Header{"Authorization": []string{"Bearer " + master.tokens["stranger"]}}
			conn, resp, err := websocket.DefaultDialer.Dial(address, header)
			if conn != nil {
				conn.Close()
			}
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			if err == nil || resp == nil || resp.StatusCode != 403 {
				t.Fatalf("unrelated user WS: response=%v err=%v", resp, err)
			}
			header.Set("Authorization", "Bearer "+master.tokens["subuser"])
			conn, _, err = websocket.DefaultDialer.Dial(address, header)
			if err != nil {
				t.Fatal("view permission denied", err)
			}
			defer conn.Close()
			conn.SetReadDeadline(time.Now().Add(3 * time.Second))
			// Drain history and immediate stats to prove view-only access works.
			for i := 0; i < 2; i++ {
				if _, _, err := conn.ReadMessage(); err != nil {
					t.Fatal(err)
				}
			}
			if err := conn.WriteJSON(WSMessage{Event: "command", Data: "stop"}); err != nil {
				t.Fatal(err)
			}
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if e, ok := err.(net.Error); ok && e.Timeout() {
						t.Fatal("unauthorized command did not close socket")
					}
					break
				}
			}
		})
	}
}

func TestConsoleRejectsCrossOriginCookieConnections(t *testing.T) {
	req := httptest.NewRequest("GET", "http://panel.example/api/servers/alpha/ws", nil)
	req.Header.Set("Origin", "https://attacker.example")
	if upgrader.CheckOrigin(req) {
		t.Fatal("cross-origin console allowed")
	}
	req.Header.Set("Origin", "https://panel.example")
	if !upgrader.CheckOrigin(req) {
		t.Fatal("same-host reverse proxy rejected")
	}
}
