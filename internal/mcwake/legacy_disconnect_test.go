package mcwake

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"
)

func optionConnection(t *testing.T, legacy bool, status func(int32) Status, join func(Join) string) (net.Conn, <-chan error) {
	t.Helper()
	server, client := net.Pipe()
	t.Cleanup(func() { server.Close(); client.Close() })
	_ = client.SetDeadline(time.Now().Add(8 * time.Second))
	done := make(chan error, 1)
	go func() { done <- HandleWithOptions(server, Options{LegacyDisconnectColors: legacy}, status, join) }()
	return client, done
}
func TestHorizonCompatibleDisconnectColors(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "semantic", true: "legacy"}[legacy], func(t *testing.T) {
			calls := 0
			c, done := optionConnection(t, legacy, nil, func(Join) string { calls++; return "&#387aff&lBrand\nSecond line\n&aReady&r plain" })
			body := append(wireString("Player"), make([]byte, 16)...)
			send(t, c, append(handshake(771, 2), wirePacket(0, body)...))
			var component textComponent
			if err := json.Unmarshal(receiveJSON(t, c), &component); err != nil {
				t.Fatal(err)
			}
			completed(t, done, false)
			closed(t, c)
			if calls != 1 {
				t.Fatal("join callback count", calls)
			}
			if legacy {
				// Horizon's json_text copies text and drops semantic styles. Literal section
				// codes must therefore exist inside this text, including each explicit line.
				if len(component.Extra) != 0 || component.Color != "" || !strings.Contains(component.Text, "§9§lBrand\n§r§9§lSecond line") || !strings.Contains(component.Text, "§aReady") {
					t.Fatal(component)
				}
				if strings.Contains(component.Text, "&#") || strings.Contains(component.Text, "§#") {
					t.Fatal("unsupported hex escape", component.Text)
				}
			} else if len(component.Extra) < 2 || component.Extra[0].Color != "#387aff" || !component.Extra[0].Bold || strings.Contains(component.Extra[0].Text, "§") {
				t.Fatal("default semantic output changed", component)
			}
		})
	}
}
func TestLegacyOptionDoesNotChangeMOTDOrWakeOnPing(t *testing.T) {
	calls := 0
	c, done := optionConnection(t, true, func(int32) Status { return Status{MOTD: "&#123abcHello", VersionName: "Sleeping"} }, func(Join) string { calls++; return "no" })
	send(t, c, append(handshake(771, 1), wirePacket(0, nil)...))
	var response struct {
		Description textComponent `json:"description"`
	}
	if err := json.Unmarshal(receiveJSON(t, c), &response); err != nil {
		t.Fatal(err)
	}
	c.Close()
	completed(t, done, false)
	if calls != 0 || response.Description.Color != "#123abc" || response.Description.Text != "Hello" {
		t.Fatal(calls, response)
	}
}
func TestLegacyDisconnectBoundsAndReset(t *testing.T) {
	got := legacyDisconnectText("&lBold&rPlain\nNext")
	if got != "§r§lBold§rPlain\n§rNext" {
		t.Fatal(got)
	}
	c, done := optionConnection(t, true, nil, func(Join) string { return strings.Repeat("&aA&bB", 1000) })
	send(t, c, append(handshake(771, 2), wirePacket(0, append(wireString("Player"), make([]byte, 16)...))...))
	raw := receiveJSON(t, c)
	if len(raw) > maxMessageBytes {
		t.Fatal("unbounded legacy output")
	}
	completed(t, done, false)
}
