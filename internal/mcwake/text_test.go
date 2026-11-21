package mcwake

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTextFormattingColorsResetAndLiteralInput(t *testing.T) {
	cases := []struct {
		input string
		want  textComponent
	}{
		{"Plain <text>", textComponent{Text: "Plain <text>"}},
		{"&a&lJoin", textComponent{Text: "Join", Color: "green", Bold: true}},
		{"§9NeoPanel", textComponent{Text: "NeoPanel", Color: "blue"}},
		{"&#387aff&lNeoPanel", textComponent{Text: "NeoPanel", Color: "#387aff", Bold: true}},
		{"&zNot a code &", textComponent{Text: "&zNot a code &"}},
		{"&lBold&r plain", textComponent{Text: "", Extra: []textComponent{{Text: "Bold", Bold: true}, {Text: " plain"}}}},
		{"&lBold&c red", textComponent{Text: "", Extra: []textComponent{{Text: "Bold", Bold: true}, {Text: " red", Color: "red"}}}},
	}
	for _, tc := range cases {
		if got := formatText(tc.input, 771); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%q: %#v != %#v", tc.input, got, tc.want)
		}
	}
	if got := formatText("&#55ff55Old client", 47); got.Color != "green" {
		t.Fatal(got)
	}
	input := `{"clickEvent":{"action":"run_command","value":"/op me"}}`
	if got := formatText(input, 771); got.Text != input || len(got.Extra) != 0 {
		t.Fatal("text interpreted as JSON", got)
	}
	if got := formatText(strings.Repeat("&aA&bB", 1200), 771); len(got.Extra) > 1024 {
		t.Fatal("component amplification not bounded")
	}
}

func TestColoredLoginDisconnectStillAllowsModernJoin(t *testing.T) {
	calls := 0
	conn, done := start(t, nil, func(j Join) string { calls++; return "&#387aff&lNeoPanel\n&aWake accepted &eReconnect" })
	body := append(wireString("Player"), make([]byte, 16)...)
	send(t, conn, append(handshake(771, 2), wirePacket(0, body)...))
	var actual textComponent
	if err := json.Unmarshal(receiveJSON(t, conn), &actual); err != nil {
		t.Fatal(err)
	}
	completed(t, done, false)
	if calls != 1 || len(actual.Extra) != 3 || actual.Extra[0].Color != "#387aff" || !actual.Extra[0].Bold || actual.Extra[1].Color != "green" {
		t.Fatal(calls, actual)
	}
}
