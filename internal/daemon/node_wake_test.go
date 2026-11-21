package daemon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func daemonWakePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
func TestServiceAppliesWakeSettingsForPanelAndWorker(t *testing.T) {
	for _, role := range []string{"panel", "node"} {
		t.Run(role, func(t *testing.T) {
			cfg := testConfig(t)
			cfg.Server.Role = role
			cfg.Security.NodeToken = "wake-fixture-machine-token"
			service := startTestService(t, cfg)
			game := &database.Server{ID: "wake-service", Name: "Sleeping fixture", Type: "paper", Port: daemonWakePort(t), NodeID: "local"}
			if err := service.db.CreateServer(game); err != nil {
				t.Fatal(err)
			}
			address := fmt.Sprintf("127.0.0.1:%d", game.Port)
			// Defaults must not bind an offline game's port.
			vacant, err := net.Listen("tcp4", address)
			if err != nil {
				t.Fatal("wake default is not disabled", err)
			}
			vacant.Close()
			wake := database.DefaultNodeWakeConfig()
			wake.Enabled = true
			wake.BindAddress = "127.0.0.1"
			wake.OfflineMOTD = "Offline test MOTD"
			data, _ := json.Marshal(wake)
			req, _ := http.NewRequest(http.MethodPut, service.URL()+"/api/cluster/wake", bytes.NewReader(data))
			req.Header.Set("Authorization", "Bearer "+cfg.Security.NodeToken)
			client := &http.Client{Timeout: 3 * time.Second}
			response, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(response.Body)
			response.Body.Close()
			if response.StatusCode != 200 {
				t.Fatal(response.StatusCode, string(body))
			}
			var conn net.Conn
			deadline := time.Now().Add(4 * time.Second)
			for time.Now().Before(deadline) {
				conn, err = net.DialTimeout("tcp4", address, 100*time.Millisecond)
				if err == nil {
					break
				}
				time.Sleep(30 * time.Millisecond)
			}
			if conn == nil {
				t.Fatal("daemon did not apply wake settings", err)
			}
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(time.Second))
			// Modern Java status handshake, then status request (never Login Start).
			handshake := []byte{0, 0x83, 0x06, 9}
			handshake = append(handshake, []byte("localhost")...)
			handshake = append(handshake, byte(game.Port>>8), byte(game.Port), 1)
			wire := append([]byte{byte(len(handshake))}, handshake...)
			wire = append(wire, 1, 0)
			if _, err := conn.Write(wire); err != nil {
				t.Fatal(err)
			}
			reader := bufio.NewReader(conn)
			size, err := binary.ReadUvarint(reader)
			if err != nil || size > 128*1024 {
				t.Fatal("status frame", size, err)
			}
			reply := make([]byte, size)
			if _, err := io.ReadFull(reader, reply); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(reply), "Offline test MOTD") || service.ActiveServerCount() != 0 {
				t.Fatal("bad offline status or ping launched game", string(reply))
			}
			if err := service.Shutdown(context.Background()); err != nil {
				t.Fatal(err)
			}
			vacant, err = net.Listen("tcp4", address)
			if err != nil {
				t.Fatal("service shutdown retained wake port", err)
			}
			vacant.Close()
		})
	}
}

func TestServiceBindFailureDoesNotStartWakeListeners(t *testing.T) {
	cfg := testConfig(t)
	occupied, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	cfg.Server.Port = occupied.Addr().(*net.TCPAddr).Port
	game := &database.Server{ID: "wake-bind-failure", Type: "paper", Port: daemonWakePort(t), NodeID: "local"}
	db, err := database.InitDB(cfg.Paths.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateServer(game); err != nil {
		t.Fatal(err)
	}
	wake := database.DefaultNodeWakeConfig()
	wake.Enabled = true
	wake.BindAddress = "127.0.0.1"
	if err := db.SetNodeWakeConfig("local", wake); err != nil {
		t.Fatal(err)
	}
	db.Close()
	service, err := Start(cfg, nil)
	if err == nil || service != nil {
		if service != nil {
			_ = service.Shutdown(context.Background())
		}
		t.Fatal("expected HTTP bind failure")
	}
	l, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", game.Port))
	if err != nil {
		t.Fatal("failed daemon startup opened wake listener", err)
	}
	l.Close()
}
