package runner

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestSteelConsoleNormalization(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"\x1b[5 q", ""}, {"\x1b[2C", ""}, {"\r\x1b[J> \x1b[38;5;15m\x1b[0m\r\x1b[2C", ""},
		{"\x1b[J  0.00s \x1b[0;1;34m[Info]\x1b[0m Starting Steel Server (1e9d4df)\r", "0.00s \x1b[0;1;34m[Info]\x1b[0m Starting Steel Server (1e9d4df)"},
		{"\r\x1b[J> \r\x1b[J  3.11s \x1b[31m[Error]\x1b[0m Address in use (os error 98)\r", "3.11s \x1b[31m[Error]\x1b[0m Address in use (os error 98)"},
		{"\x1b]0;window title\x07\x1b]8;;https://example.test\x1b\\message\x1b]8;;\x1b\\", "message"},
		{"old progress\rnew progress\r", "new progress"},
		{"\x1b[38;2;12;34;56mRGB\x1b[0m", "\x1b[38;2;12;34;56mRGB\x1b[0m"},
	} {
		got, keep := normalizeSteelConsoleLine(tc.raw)
		if got != tc.want || keep != (tc.want != "") {
			t.Errorf("normalize %q: got %q/%v, want %q", tc.raw, got, keep, tc.want)
		}
	}
}

func TestSteelStreamDropsPromptsPreservesColorAndDiagnosesBind(t *testing.T) {
	p := NewProcessInstance(&database.Server{ID: "steel", Type: "steel", Port: 25565}, t.TempDir(), time.Second, nil, nil)
	p.status, p.runID = "starting", 1
	source := "\x1b[5 q\n\x1b[2C\n\r\x1b[J> \x1b[0m\n" +
		"\r\x1b[J 0.00s \x1b[0;1;34m[Info]\x1b[0m Starting Steel Server\n" +
		"\r\x1b[J 3.11s \x1b[31m[Error]\x1b[0m Server startup failed: failed to bind to server port 25565: Address in use (os error 98)\n"
	p.streamOutput(strings.NewReader(source), 1)
	logs := strings.Join(p.logBuffer, "\n")
	if len(p.logBuffer) != 6 {
		t.Fatalf("unexpected records (%d): %s", len(p.logBuffer), logs)
	}
	if strings.Contains(logs, "\x1b[J") || strings.Contains(logs, "\x1b[2C") || strings.Contains(logs, "\x1b[5 q") || strings.Contains(logs, "\r") {
		t.Fatal("leaked terminal controls", logs)
	}
	if !strings.Contains(logs, "\x1b[0;1;34m[Info]") || !strings.Contains(logs, "\x1b[31m[Error]") {
		t.Fatal("lost colors", logs)
	}
	if !strings.Contains(logs, "sudo ss -tanp '( sport = :25565 )'") || !strings.Contains(logs, "TIME_WAIT") {
		t.Fatal("missing port diagnosis", logs)
	}
	if p.GetStatus() != "starting" {
		t.Fatal("bind failure marked ready")
	}
}

func TestSteelCleanedReadinessAndPlayers(t *testing.T) {
	p := NewProcessInstance(&database.Server{ID: "steel", Type: "steel"}, t.TempDir(), time.Second, nil, nil)
	p.status, p.runID = "starting", 1
	p.streamOutput(strings.NewReader("\r\x1b[J 0.94s \x1b[34m[Info]\x1b[0m Started Steel Server\n\r\x1b[J> \r\x1b[J 1.00s [Info] Steve joined the game\n"), 1)
	if p.GetStatus() != "online" || p.GetPlayerCount() != 1 {
		t.Fatal("cleaning broke readiness or player tracking", p.logBuffer)
	}
}

func TestStartupFailureMessagesDistinguishExitFromTimeout(t *testing.T) {
	for _, timedOut := range []bool{false, true} {
		dir := t.TempDir()
		db, err := database.InitDB(filepath.Join(dir, "panel.db"))
		if err != nil {
			t.Fatal(err)
		}
		m := NewServerManager(filepath.Join(dir, "servers"), db)
		srv := &database.Server{ID: "steel", Name: "Steel", Type: "steel", AutoRestart: true}
		if err := db.CreateServer(srv); err != nil {
			t.Fatal(err)
		}
		inst := m.getOrCreateInstance(srv)
		m.startupFailures[srv.ID] = MaxStartupAttempts - 1
		m.handleProcessExit(srv.ID, true, false, timedOut, time.Second)
		logs := strings.Join(inst.logBuffer, "\n")
		m.ShutdownAll()
		db.Close()
		want := "STARTUP FAILED"
		if timedOut {
			want = "STARTUP TIMED OUT"
		}
		if !strings.Contains(logs, want) || strings.Contains(logs, "Java settings") || (!timedOut && strings.Contains(logs, "TIMED OUT")) {
			t.Fatal("misleading startup error", logs)
		}
	}
}
