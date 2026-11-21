package runner

import (
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func idleTrackingFixture(t *testing.T) *ProcessInstance {
	t.Helper()
	p := NewProcessInstance(&database.Server{ID: "idle", Type: "paper"}, t.TempDir(), time.Second, nil, nil)
	p.status = "online"
	p.runID = 7
	p.zeroPlayersSince = time.Now().Add(-20 * time.Minute)
	return p
}

func TestIdleEmptyPollingPreservesElapsedTime(t *testing.T) {
	p := idleTrackingFixture(t)
	since := p.zeroPlayersSince
	for i := 0; i < 10; i++ {
		p.trackPlayerLine("[12:00:00] [Server thread/INFO]: There are 0 of a max of 20 players online:", 7)
	}
	if !p.zeroPlayersSince.Equal(since) {
		t.Fatal("empty list polling restarted the idle clock")
	}
	if duration, idle := p.GetIdleDuration(); !idle || duration < 20*time.Minute {
		t.Fatal("lost idle duration", duration, idle)
	}
}

func TestIdleDuplicateDisconnectDoesNotRestartClock(t *testing.T) {
	p := idleTrackingFixture(t)
	since := p.zeroPlayersSince
	for _, line := range []string{"Player lost connection: Disconnected", "Player left the game", "[connected player] Player has disconnected", "[Player] -> UpstreamBridge has disconnected"} {
		p.trackPlayerLine(line, 7)
	}
	if !p.zeroPlayersSince.Equal(since) {
		t.Fatal("duplicate/unknown disconnect restarted the idle clock")
	}
}

func TestIdlePositiveAnonymousCountIsNotEmpty(t *testing.T) {
	p := idleTrackingFixture(t)
	p.trackPlayerLine("There are 2 of a max of 20 players online:", 7)
	if _, idle := p.GetIdleDuration(); idle {
		t.Fatal("positive player count without names treated as empty")
	}
	if count := p.GetPlayerCount(); count < 2 {
		t.Fatal("anonymous players disappeared", count)
	}
	p.trackPlayerLine("Unknown left the game", 7)
	if _, idle := p.GetIdleDuration(); idle {
		t.Fatal("unresolved player count cleared by unrelated leave")
	}
	p.trackPlayerLine("There are 0 of a max of 20 players online:", 7)
	if duration, idle := p.GetIdleDuration(); !idle || duration > time.Second {
		t.Fatal("new empty interval not started", duration, idle)
	}
}

func TestIdlePlayersAndRunGenerationResetClock(t *testing.T) {
	p := idleTrackingFixture(t)
	p.trackPlayerLine("Player joined the game", 7)
	if _, idle := p.GetIdleDuration(); idle {
		t.Fatal("joined player considered idle")
	}
	p.trackPlayerLine("There are 0 of a max of 20 players online:", 6)
	if p.GetPlayerCount() != 1 {
		t.Fatal("stale output changed occupancy")
	}
	p.trackPlayerLine("Player left the game", 7)
	if duration, idle := p.GetIdleDuration(); !idle || duration > time.Second {
		t.Fatal("leave failed to start new interval", duration, idle)
	}
	p.trackPlayerLine("There are 2 of a max of 20 players online: One", 7)
	if _, idle := p.GetIdleDuration(); idle {
		t.Fatal("partial list treated as empty")
	}
	if p.GetPlayerCount() != 2 {
		t.Fatal("partial list lost unnamed player")
	}
}
