package runner

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWakeListenerUsesSavedHorizonCompatibility(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "graceful")
	cfg.LegacyDisconnectColors = true
	cfg.StartingMessage = "&#387aff&lBrand\n&aPlease reconnect"
	if err := m.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	raw := wakeRequest(t, s, true)
	var component struct {
		Text  string `json:"text"`
		Extra []any  `json:"extra"`
	}
	if err := json.Unmarshal([]byte(raw), &component); err != nil {
		t.Fatal(err)
	}
	if len(component.Extra) != 0 || !strings.Contains(component.Text, "§9§lBrand") || !strings.Contains(component.Text, "§aPlease reconnect") {
		t.Fatal("saved compatibility flag did not reach socket response", raw)
	}
	p, _ := m.GetInstance(s.ID)
	waitProcessStatus(t, p, "online")
}
