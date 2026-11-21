package runner

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWakeIconConfinedAndFallback(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 64, 64))); err != nil {
		t.Fatal(err)
	}
	cfg.IconDataURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(b.Bytes())
	if got := m.wakeIcon(s.ID, cfg); got != cfg.IconDataURL {
		t.Fatal("missing file did not use fallback")
	}
	iconPath := filepath.Join(m.GetWorkingDir(s.ID), "server-icon.png")
	if err := os.WriteFile(iconPath, b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	expected := cfg.IconDataURL
	cfg.IconDataURL = ""
	if got := m.wakeIcon(s.ID, cfg); got != expected {
		t.Fatal("valid server icon ignored")
	}
	cfg.UseServerIcon = false
	if got := m.wakeIcon(s.ID, cfg); got != "" {
		t.Fatal("server icon not disabled")
	}
	cfg.UseServerIcon = true
	if err := os.WriteFile(iconPath, []byte("not PNG"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := m.wakeIcon(s.ID, cfg); got != "" {
		t.Fatal("invalid icon accepted")
	}
	if got := m.wakeIcon("../escape", cfg); got != "" {
		t.Fatal("path escaped server root")
	}
}

func TestWakeAutoRetryHandoffPreservesFailureCount(t *testing.T) {
	m, s, _ := wakeFixture(t)
	s.AutoRestart = true
	if err := m.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "wake")
	t.Setenv("NEOPANEL_WAKE_TEST_ADDRESS", wakeAddress(s))
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	cancel := make(chan struct{})
	m.mu.Lock()
	m.retryCancels[s.ID] = cancel
	m.startupFailures[s.ID] = 1
	m.mu.Unlock()
	m.waitAndRetryStart(s.ID, 2, cancel)
	p, _ := m.GetInstance(s.ID)
	waitProcessStatus(t, p, "online")
	m.mu.RLock()
	failures := m.startupFailures[s.ID]
	m.mu.RUnlock()
	if failures != 1 || wakeEntry(m, s.ID) != nil {
		t.Fatal("retry reset failures or kept wake port", failures)
	}
}

func TestWakeOccupiedPortAndConfigurationRefresh(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	other, err := net.Listen("tcp4", wakeAddress(s))
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	m.StartWakeListeners()
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("hijacked occupied port")
	}
	other.Close()
	m.reconcileWakeListeners()
	old := wakeEntry(m, s.ID)
	if old == nil {
		t.Fatal("did not retry available port")
	}
	cfg.OfflineMOTD = "&bChanged {server}"
	if err := m.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	updated := wakeEntry(m, s.ID)
	if updated == nil || updated == old || updated.config.OfflineMOTD != cfg.OfflineMOTD {
		t.Fatal("settings not refreshed")
	}
	// A connection admitted before a configuration change cannot wake from its
	// stale generation, even if the new configuration is enabled too.
	if m.reserveWake(old) {
		t.Fatal("stale generation admitted")
	}
	m.wake.mu.Lock()
	m.wake.cooldown[s.ID] = time.Now().Add(time.Minute)
	m.wake.mu.Unlock()
	if m.reserveWake(updated) {
		t.Fatal("cooldown bypass")
	}
}

func TestWakeReconcilePreservesAcceptedRequestDuringGateContention(t *testing.T) {
	m, s, _ := wakeFixture(t)
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "wake")
	t.Setenv("NEOPANEL_WAKE_TEST_ADDRESS", wakeAddress(s))
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	e := wakeEntry(m, s.ID)
	if !m.reserveWake(e) {
		t.Fatal("wake not accepted")
	}
	// Model launchWake owning the operation gate while reconciliation owns
	// launchMu. It must not delete the accepted listener generation.
	release, err := m.beginOperation(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	m.reconcileWakeListeners()
	preserved := wakeEntry(m, s.ID) == e
	release()
	if !preserved {
		t.Fatal("reconcile cancelled an accepted wake")
	}
	m.launchWake(e)
	p, _ := m.GetInstance(s.ID)
	waitProcessStatus(t, p, "online")
}
