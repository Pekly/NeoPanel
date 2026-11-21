package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func wakeFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
func wakeFixture(t *testing.T) (*ServerManager, *database.Server, database.NodeWakeConfig) {
	t.Helper()
	db, base, s, _ := restartFixture(t)
	s.Port = wakeFreePort(t)
	if err := db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(base, db)
	t.Cleanup(m.ShutdownAll)
	m.getOrCreateInstance(s).SetIsolationConfig("process", 512)
	cfg := database.DefaultNodeWakeConfig()
	cfg.Enabled = true
	cfg.BindAddress = "127.0.0.1"
	cfg.CooldownSeconds = 10
	if err := db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	// Existing protocol fixtures represent a previously completed trusted idle stop.
	if err := db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 15}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetServerIdleStopped(s.ID, true); err != nil {
		t.Fatal(err)
	}
	return m, s, cfg
}
func wakeEntry(m *ServerManager, id string) *wakeListener {
	m.wake.mu.Lock()
	defer m.wake.mu.Unlock()
	return m.wake.listeners[id]
}
func wakeWait(t *testing.T, condition func() bool) {
	t.Helper()
	end := time.Now().Add(5 * time.Second)
	for time.Now().Before(end) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("wake condition timed out")
}
func wakeAddress(s *database.Server) string { return fmt.Sprintf("127.0.0.1:%d", s.Port) }
func wakeVar(n uint64) []byte               { var b [10]byte; k := binary.PutUvarint(b[:], n); return b[:k] }
func wakeString(s string) []byte            { return append(wakeVar(uint64(len(s))), []byte(s)...) }
func wakeFrame(b []byte) []byte             { return append(wakeVar(uint64(len(b))), b...) }
func wakeRequest(t *testing.T, s *database.Server, login bool) string {
	t.Helper()
	c, err := net.DialTimeout("tcp4", wakeAddress(s), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(3 * time.Second))
	state := byte(1)
	if login {
		state = 2
	}
	h := append([]byte{0}, wakeVar(771)...)
	h = append(h, wakeString("localhost")...)
	h = append(h, byte(s.Port>>8), byte(s.Port), state)
	packet := []byte{0}
	if login {
		packet = append(packet, wakeString("Player_1")...)
		packet = append(packet, make([]byte, 16)...)
	}
	if _, err := c.Write(append(wakeFrame(h), wakeFrame(packet)...)); err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(c)
	n, err := binary.ReadUvarint(r)
	if err != nil {
		t.Fatal(err)
	}
	if n > 128*1024 {
		t.Fatal("oversized response")
	}
	body := make([]byte, n)
	if _, err := io.ReadFull(r, body); err != nil {
		t.Fatal(err)
	}
	b := bytes.NewReader(body)
	id, _ := binary.ReadUvarint(b)
	size, err := binary.ReadUvarint(b)
	if err != nil || id != 0 || size > uint64(b.Len()) {
		t.Fatal("bad response")
	}
	text := make([]byte, size)
	_, _ = io.ReadFull(b, text)
	return string(text)
}

func TestWakePingJoinPortHandoffAndCooldown(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "wake")
	t.Setenv("NEOPANEL_WAKE_TEST_ADDRESS", wakeAddress(s))
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	response := wakeRequest(t, s, false)
	var status struct {
		Description json.RawMessage `json:"description"`
		Version     struct {
			Protocol int `json:"protocol"`
		} `json:"version"`
	}
	if err := json.Unmarshal([]byte(response), &status); err != nil || !strings.Contains(string(status.Description), s.Name) || status.Version.Protocol != -1 || !strings.Contains(string(status.Description), `"color"`) {
		t.Fatal(response, err)
	}
	if m.GetStatus(s.ID) != "offline" {
		t.Fatal("ping started server")
	}
	message := wakeRequest(t, s, true)
	if !strings.Contains(message, "reconnect") {
		t.Fatal(message)
	}
	inst, _ := m.GetInstance(s.ID)
	waitProcessStatus(t, inst, "online")
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("wake still owns game port")
	}
	addr, err := os.ReadFile(filepath.Join(m.GetWorkingDir(s.ID), "listener.addr"))
	if err != nil || string(addr) != wakeAddress(s) {
		t.Fatal("game did not acquire port", string(addr), err)
	}
	if err := m.Stop(s.ID, true); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("manual stop reopened listener")
	}
	// Simulate a later completed idle stop to verify cooldown retention.
	if err := m.db.SetServerIdleStopped(s.ID, true); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	if e := wakeEntry(m, s.ID); e == nil || m.reserveWake(e) {
		t.Fatal("cooldown not retained after stop")
	}
	cfg.Enabled = false
	if err := m.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("disable left listener")
	}
	l, err := net.Listen("tcp4", wakeAddress(s))
	if err != nil {
		t.Fatal("disabled listener kept port", err)
	}
	l.Close()
}

func TestWakeManualAndRecoveryStartReleasePort(t *testing.T) {
	for _, recovery := range []bool{false, true} {
		t.Run(fmt.Sprint(recovery), func(t *testing.T) {
			m, s, _ := wakeFixture(t)
			t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "wake")
			t.Setenv("NEOPANEL_WAKE_TEST_ADDRESS", wakeAddress(s))
			m.StartWakeListeners()
			wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
			var err error
			if recovery {
				err = m.db.SaveServerResumeState(m.baseDir, []string{s.ID})
				if err == nil {
					err = m.RestoreRunningServers()
				}
			} else {
				err = m.Start(s)
			}
			if err != nil {
				t.Fatal(err)
			}
			p, _ := m.GetInstance(s.ID)
			waitProcessStatus(t, p, "online")
			if wakeEntry(m, s.ID) != nil {
				t.Fatal("listener survived start")
			}
		})
	}
}

func TestWakeInventoryAndOperationGuards(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	e := wakeEntry(m, s.ID)
	release, err := m.BeginServerMaintenance(context.Background(), s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if wakeEntry(m, s.ID) != nil || m.reserveWake(e) {
		t.Fatal("maintenance allows wake")
	}
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("rebound during maintenance")
	}
	release()
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("maintenance reopened idle authority")
	}
	// Remaining guard checks use a new trusted idle-stop fixture.
	if err := m.db.SetServerIdleStopped(s.ID, true); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	m.mu.Lock()
	m.startupHalted[s.ID] = true
	m.mu.Unlock()
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("halted listener")
	}
	m.mu.Lock()
	m.startupHalted[s.ID] = false
	m.mu.Unlock()
	m.restoreBlocked.Store(s.ID, ErrRestoreRecovery)
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("blocked listener")
	}
	m.restoreBlocked.Delete(s.ID)
	oldPort := s.Port
	s.Port = wakeFreePort(t)
	if err := m.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	if e := wakeEntry(m, s.ID); e == nil || e.server.Port != s.Port {
		t.Fatal("port change not applied")
	}
	l, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", oldPort))
	if err != nil {
		t.Fatal(err)
	}
	l.Close()
	// A reserved login cannot start after a disable or a deleted inventory row.
	e = wakeEntry(m, s.ID)
	if !m.reserveWake(e) {
		t.Fatal("reservation refused")
	}
	cfg.Enabled = false
	if err := m.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	m.launchWake(e)
	if m.GetStatus(s.ID) != "offline" {
		t.Fatal("disabled wake launched")
	}
	cfg.Enabled = true
	_ = m.db.SetNodeWakeConfig("local", cfg)
	m.reconcileWakeListeners()
	if err := m.DeleteServer(context.Background(), s.ID, func() error { return m.db.DeleteServer(s.ID) }); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("deleted listener returned")
	}
}

func TestWakeDisabledDefaultAndLocalJavaInventory(t *testing.T) {
	m, s, _ := wakeFixture(t)
	cfg := database.DefaultNodeWakeConfig()
	_ = m.db.SetNodeWakeConfig("local", cfg)
	m.StartWakeListeners()
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("default enabled")
	}
	for _, tc := range []struct {
		node, kind string
		want       bool
	}{{"local", "paper", true}, {"", "steel", true}, {"worker", "paper", false}, {"local", "bedrock", false}, {"local", "unknown", false}} {
		copy := *s
		copy.NodeID = tc.node
		copy.Type = tc.kind
		if localWakeServer(&copy) != tc.want {
			t.Fatalf("inventory %+v", tc)
		}
	}
	cfg.Enabled = true
	_ = m.db.SetNodeWakeConfig("local", cfg)
	conflict := *s
	conflict.ID = "port-conflict"
	if err := m.db.CreateServer(&conflict); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil || wakeEntry(m, conflict.ID) != nil {
		t.Fatal("ambiguous port bound")
	}
}

func TestWakeShutdownClosesConnectionsAndDoesNotSnapshotListeners(t *testing.T) {
	m, s, _ := wakeFixture(t)
	if err := m.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	c, err := net.Dial("tcp4", wakeAddress(s))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	wakeWait(t, func() bool {
		m.wake.mu.Lock()
		defer m.wake.mu.Unlock()
		return len(m.wake.listeners[s.ID].connections) > 0
	})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); m.ShutdownAll() }()
	}
	wg.Wait()
	_ = c.SetReadDeadline(time.Now().Add(time.Second))
	var b [1]byte
	if _, err := c.Read(b[:]); err == nil {
		t.Fatal("connection survived shutdown")
	}
	if ids, err := m.db.TakeServerResumeState(m.baseDir); err != nil || len(ids) != 0 {
		t.Fatal("listener became recovery intent", ids, err)
	}
	m.StartWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("reopened after shutdown")
	}
}

func TestWakeReservationDedupeFailureAndStaleSettings(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	e := wakeEntry(m, s.ID)
	var wg sync.WaitGroup
	var mu sync.Mutex
	accepted := 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if m.reserveWake(e) {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if accepted != 1 {
		t.Fatal("duplicate wake", accepted)
	}
	// Ensure synchronous launch failure halts future external retries.
	s.JavaPath = filepath.Join(t.TempDir(), "missing-java")
	if err := m.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	m.launchWake(e)
	m.mu.RLock()
	halted := m.startupHalted[s.ID]
	m.mu.RUnlock()
	if !halted {
		t.Fatal("failed external launch not halted")
	}
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("failed launch reopened")
	}
	_ = cfg
}

func TestWakeConnectionRateLimitAndMalformedLogin(t *testing.T) {
	m, s, _ := wakeFixture(t)
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	for i := 0; i < 14; i++ {
		c, err := net.Dial("tcp4", wakeAddress(s))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = c.Write([]byte{1, 255})
		_ = c.SetReadDeadline(time.Now().Add(time.Second))
		_, _ = io.Copy(io.Discard, c)
		c.Close()
	}
	if m.GetStatus(s.ID) != "offline" {
		t.Fatal("malformed packets woke server")
	}
	m.wake.mu.Lock()
	budget := m.wake.ips["127.0.0.1"]
	slots := len(m.wake.slots)
	m.wake.mu.Unlock()
	if budget.count != 12 || slots > 1 {
		t.Fatal("rate budget", budget, slots)
	}
	if got := wakeText("&a{server}\n{version} &z", *s); got != "§a"+s.Name+"\n"+s.Version+" &z" {
		t.Fatal(got)
	}
}
