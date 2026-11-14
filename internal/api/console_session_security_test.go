package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// The fake node deliberately never revalidates its machine token: only the
// master's original-client validation can revoke these proxied connections.
func consoleSessionEndpoint(t *testing.T, f *reliabilityFixture, proxy bool) string {
	t.Helper()
	if proxy {
		node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer machine-token" {
				http.Error(w, "bad machine token", 401)
				return
			}
			if r.URL.Path == "/api/auth/me" {
				w.Header().Set("X-NeoPanel-User-ID", r.Header.Get("X-NeoPanel-User-ID"))
				w.Header().Set("X-NeoPanel-Permission-Policy", "2")
				w.Write([]byte(`{}`))
				return
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()
			conn.WriteJSON(WSMessage{Event: "history", Data: "[]"})
			conn.WriteJSON(WSMessage{Event: "stats"})
			for {
				var msg WSMessage
				if conn.ReadJSON(&msg) != nil {
					return
				}
				if msg.Event == "command" {
					msg.Event = "command_received"
				} else {
					msg.Event = "pong"
				}
				if conn.WriteJSON(msg) != nil {
					return
				}
			}
		}))
		t.Cleanup(node.Close)
		if err := f.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "remote", BaseURL: node.URL, APIToken: "machine-token"}); err != nil {
			t.Fatal(err)
		}
	}
	server := httptest.NewServer(f.router)
	t.Cleanup(server.Close)
	path := "/api/servers/alpha/ws"
	if proxy {
		path = "/api/nodes/remote/ws/alpha"
	}
	return "ws" + strings.TrimPrefix(server.URL, "http") + path
}

func consoleSessionDial(t *testing.T, address, token, transport string) *websocket.Conn {
	t.Helper()
	header := http.Header{}
	switch transport {
	case "cookie":
		header.Set("Cookie", "neopanel_token="+token)
	case "query":
		address += "?token=" + url.QueryEscape(token)
	default:
		header.Set("Authorization", "Bearer "+token)
	}
	conn, resp, err := websocket.DefaultDialer.Dial(address, header)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for i := 0; i < 2; i++ {
		if _, _, err := conn.ReadMessage(); err != nil {
			t.Fatal("initial console messages", err)
		}
	}
	return conn
}

func consoleSessionClosed(t *testing.T, conn *websocket.Conn, limit time.Duration) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(limit))
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				t.Fatal("revoked console remained open")
			}
			return
		}
		if msg.Event == "pong" || msg.Event == "command_received" {
			t.Fatal("message processed after credential revocation", msg.Event)
		}
	}
}

func consoleSessionKey(t *testing.T, f *reliabilityFixture) (string, int) {
	t.Helper()
	raw := "neo_console-session-test"
	sum := sha256.Sum256([]byte(raw))
	key, err := f.db.AddAPIKey(f.users["admin"].ID, "console test", hex.EncodeToString(sum[:]), "neo_console")
	if err != nil {
		t.Fatal(err)
	}
	return raw, key.ID
}

func TestConsoleSessionRevocationLocalAndProxy(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		kind := "local"
		if proxy {
			kind = "proxy"
		}
		for _, mode := range []string{"password", "reset", "demotion", "key-revocation", "expiry"} {
			for _, command := range []bool{false, true} {
				action := "idle"
				if command {
					action = "command"
				}
				t.Run(kind+"/"+mode+"/"+action, func(t *testing.T) {
					f := newReliabilityFixture(t)
					address := consoleSessionEndpoint(t, f, proxy)
					token := f.tokens["admin"]
					keyID := 0
					if mode == "key-revocation" {
						token, keyID = consoleSessionKey(t, f)
					}
					expires := time.Now().Add(2 * time.Second).Truncate(time.Second)
					if mode == "expiry" {
						claims := JWTClaims{UserID: f.users["admin"].ID, Username: "admin", Role: "admin", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expires)}}
						var err error
						token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(f.cfg.Security.JWTSecret))
						if err != nil {
							t.Fatal(err)
						}
					}
					transport := "header"
					if mode == "reset" {
						transport = "cookie"
					}
					if mode == "password" {
						transport = "query"
					}
					conn := consoleSessionDial(t, address, token, transport)
					var err error
					switch mode {
					case "password":
						err = f.db.UpdateUserPassword(f.users["admin"].ID, "changed")
					case "reset":
						err = f.db.CreatePasswordResetToken(f.users["admin"].ID, "reset-console", time.Now().Add(time.Hour))
						if err == nil {
							err = f.db.ResetUserPassword("reset-console", "changed")
						}
					case "demotion":
						err = f.db.UpdateUserRole(f.users["admin"].ID, "user")
					case "key-revocation":
						err = f.db.DeleteAPIKey(f.users["admin"].ID, keyID)
					case "expiry":
						time.Sleep(time.Until(expires) + 20*time.Millisecond)
					}
					if err != nil {
						t.Fatal(err)
					}
					if command {
						// A queued ping proves the command loop did not simply ignore the revoked
						// command and keep the socket usable; the fake node also reports any relay.
						_ = conn.WriteJSON(WSMessage{Event: "command", Data: "stop"})
						_ = conn.WriteJSON(WSMessage{Event: "ping"})
					}
					consoleSessionClosed(t, conn, 2500*time.Millisecond)
				})
			}
		}
	}
}

func TestConsoleSessionPendingTokensRejected(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		f := newReliabilityFixture(t)
		address := consoleSessionEndpoint(t, f, proxy)
		token, err := f.router.auth.Generate2FATempToken(f.users["admin"])
		if err != nil {
			t.Fatal(err)
		}
		conn, resp, err := websocket.DefaultDialer.Dial(address, http.Header{"Authorization": []string{"Bearer " + token}})
		if conn != nil {
			conn.Close()
		}
		if resp != nil {
			resp.Body.Close()
		}
		if err == nil || resp == nil || resp.StatusCode != 401 {
			t.Fatalf("pending token opened console: proxy=%v response=%v error=%v", proxy, resp, err)
		}
	}
}

func TestConsoleSessionMachineAndAPIKeysRemainIndependentOfPasswords(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		for _, credential := range []string{"node", "api-key"} {
			name := credential + "/local"
			if proxy {
				name = credential + "/proxy"
			}
			t.Run(name, func(t *testing.T) {
				f := newReliabilityFixture(t)
				address := consoleSessionEndpoint(t, f, proxy)
				token := f.cfg.Security.NodeToken
				if credential == "api-key" {
					token, _ = consoleSessionKey(t, f)
				}
				conn := consoleSessionDial(t, address, token, "header")
				if err := f.db.UpdateUserPassword(f.users["admin"].ID, "changed"); err != nil {
					t.Fatal(err)
				}
				time.Sleep(1100 * time.Millisecond)
				if err := conn.WriteJSON(WSMessage{Event: "ping"}); err != nil {
					t.Fatal(err)
				}
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				for {
					var msg WSMessage
					if err := conn.ReadJSON(&msg); err != nil {
						t.Fatal("independent credential revoked", err)
					}
					if msg.Event == "pong" {
						break
					}
				}
			})
		}
	}
}
