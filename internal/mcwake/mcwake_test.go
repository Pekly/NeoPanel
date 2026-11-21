package mcwake

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Wire helpers deliberately use encoding/binary, not the production codec.
func wireVar(n uint64) []byte    { return binary.AppendUvarint(nil, n) }
func wireString(s string) []byte { return append(wireVar(uint64(len(s))), s...) }
func wirePacket(id byte, body []byte) []byte {
	return append(wireVar(uint64(len(body)+1)), append([]byte{id}, body...)...)
}
func handshakeBody(protocol, state int, host string) []byte {
	b := append(wireVar(uint64(protocol)), wireString(host)...)
	return append(b, 0x63, 0xdd, byte(state))
}
func handshake(protocol, state int) []byte {
	return wirePacket(0, handshakeBody(protocol, state, "127.0.0.1\x00ignored-routing-data"))
}
func loginPacket(name string, tail []byte) []byte {
	return wirePacket(0, append(wireString(name), tail...))
}

type byteReader struct{ io.Reader }

func (r byteReader) ReadByte() (byte, error) {
	var b [1]byte
	_, err := io.ReadFull(r.Reader, b[:])
	return b[0], err
}
func receive(t *testing.T, conn net.Conn) (byte, []byte) {
	t.Helper()
	n, err := binary.ReadUvarint(byteReader{conn})
	if err != nil || n < 1 || n > maxStatusBytes {
		t.Fatalf("response frame length = %d, err = %v", n, err)
	}
	b := make([]byte, int(n))
	if _, err := io.ReadFull(conn, b); err != nil {
		t.Fatal(err)
	}
	return b[0], b[1:]
}
func receiveJSON(t *testing.T, conn net.Conn) []byte {
	t.Helper()
	id, body := receive(t, conn)
	if id != 0 {
		t.Fatalf("response packet ID = %d", id)
	}
	n, used := binary.Uvarint(body)
	if used <= 0 || n != uint64(len(body)-used) {
		t.Fatalf("expected exactly one length-prefixed JSON string: %x", body)
	}
	if !json.Valid(body[used:]) {
		t.Fatalf("invalid JSON: %q", body[used:])
	}
	return body[used:]
}
func send(t *testing.T, conn net.Conn, b []byte) {
	t.Helper()
	if _, err := conn.Write(b); err != nil {
		t.Fatal(err)
	}
}
func start(t *testing.T, status func(int32) Status, join func(Join) string) (net.Conn, <-chan error) {
	t.Helper()
	server, client := net.Pipe()
	t.Cleanup(func() { client.Close(); server.Close() })
	if err := client.SetDeadline(time.Now().Add(8 * time.Second)); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- Handle(server, status, join) }()
	return client, done
}
func completed(t *testing.T, done <-chan error, wantError bool) {
	t.Helper()
	select {
	case err := <-done:
		if (err != nil) != wantError {
			t.Fatalf("Handle error = %v, want error %v", err, wantError)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("Handle did not finish")
	}
}
func closed(t *testing.T, conn net.Conn) {
	t.Helper()
	var b [1]byte
	if n, err := conn.Read(b[:]); n != 0 || err != io.EOF {
		t.Fatalf("connection not closed/no-extra-output violated: n=%d err=%v", n, err)
	}
}

func TestStatusAndPing(t *testing.T) {
	for _, icon := range []string{"", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII="} {
		t.Run(fmt.Sprintf("icon=%t", icon != ""), func(t *testing.T) {
			var wakes, statuses atomic.Int32
			motd := "§aOffline\n\"text\":\"injection\" <&>"
			conn, done := start(t, func(protocol int32) Status {
				statuses.Add(1)
				if protocol != 771 {
					t.Errorf("protocol=%d", protocol)
				}
				return Status{MOTD: motd, VersionName: "NeoPanel", MaxPlayers: 20, IconDataURL: icon}
			}, func(Join) string { wakes.Add(1); return "no" })
			send(t, conn, append(handshake(771, 1), wirePacket(0, nil)...))
			var actual map[string]any
			if err := json.Unmarshal(receiveJSON(t, conn), &actual); err != nil {
				t.Fatal(err)
			}
			want := map[string]any{
				"version":     map[string]any{"name": "NeoPanel", "protocol": float64(-1)},
				"players":     map[string]any{"max": float64(20), "online": float64(0)},
				"description": map[string]any{"text": strings.TrimPrefix(motd, "§a"), "color": "green"},
			}
			if icon != "" {
				want["favicon"] = icon
			}
			if !reflect.DeepEqual(actual, want) {
				t.Fatalf("status=%#v, want %#v", actual, want)
			}
			ping := []byte{0xff, 0, 1, 2, 3, 4, 5, 6}
			send(t, conn, wirePacket(1, ping))
			id, got := receive(t, conn)
			if id != 1 || !bytes.Equal(got, ping) {
				t.Fatalf("ping %d %x", id, got)
			}
			completed(t, done, false)
			closed(t, conn)
			if wakes.Load() != 0 || statuses.Load() != 1 {
				t.Fatalf("calls status=%d join=%d", statuses.Load(), wakes.Load())
			}
		})
	}
}

func TestStatusOptionalPingAndMalformedPing(t *testing.T) {
	for _, tc := range []struct {
		name string
		ping []byte
		bad  bool
	}{
		{"EOF", nil, false}, {"partial length", []byte{0x80}, true},
		{"missing body", []byte{9}, true}, {"partial body", []byte{9, 1, 0}, true},
		{"wrong ID", wirePacket(0, make([]byte, 8)), true},
		{"short", wirePacket(1, make([]byte, 7)), true},
		{"extra", wirePacket(1, make([]byte, 9)), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var wakes atomic.Int32
			conn, done := start(t, func(int32) Status { return Status{} }, func(Join) string { wakes.Add(1); return "" })
			send(t, conn, append(handshake(47, 1), wirePacket(0, nil)...))
			receiveJSON(t, conn)
			if len(tc.ping) != 0 {
				send(t, conn, tc.ping)
			}
			conn.Close()
			completed(t, done, tc.bad)
			if wakes.Load() != 0 {
				t.Fatal("status woke server")
			}
		})
	}
}

func signatureTail() []byte {
	b := append([]byte{1}, make([]byte, 8)...)
	b = append(b, wireString("public-key")...)
	return append(b, wireString("signature")...)
}

func TestLoginVersions(t *testing.T) {
	uuid := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	cases := []struct {
		name            string
		protocol, state int
		tail            []byte
	}{
		{"1.7.2", 4, 2, nil}, {"1.8", 47, 2, nil}, {"1.18.2", 758, 2, nil},
		{"1.19 absent", 759, 2, []byte{0}}, {"1.19 signature", 759, 2, signatureTail()},
		{"1.19.2 absent", 760, 2, []byte{0, 0}},
		{"1.19.2 signed UUID", 760, 2, append(append(signatureTail(), 1), uuid...)},
		{"1.19.2 UUID", 760, 2, append([]byte{0, 1}, uuid...)},
		{"1.19.2 signed", 760, 2, append(signatureTail(), 0)},
		{"1.19.3 absent", 761, 2, []byte{0}}, {"1.19.3 UUID", 761, 2, append([]byte{1}, uuid...)},
		{"1.19.4 absent", 762, 2, []byte{0}}, {"1.19.4 UUID", 762, 2, append([]byte{1}, uuid...)},
		{"1.20.1 absent", 763, 2, []byte{0}}, {"1.20.1 UUID", 763, 2, append([]byte{1}, uuid...)},
		{"1.20.2", 764, 2, uuid}, {"1.21.6", 771, 2, uuid},
		{"1.20.5 transfer", 766, 3, uuid}, {"1.21.6 transfer", 771, 3, uuid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var wakes, statuses atomic.Int32
			message := "§eStarting…\nReconnect! \"text\":\"injected\""
			conn, done := start(t, func(int32) Status { statuses.Add(1); return Status{} }, func(j Join) string {
				wakes.Add(1)
				if j.Protocol != int32(tc.protocol) || j.Username != "Player_123" {
					t.Errorf("join=%+v", j)
				}
				return message
			})
			send(t, conn, append(handshake(tc.protocol, tc.state), loginPacket("Player_123", tc.tail)...))
			var component map[string]string
			if err := json.Unmarshal(receiveJSON(t, conn), &component); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(component, map[string]string{"text": strings.TrimPrefix(message, "§e"), "color": "yellow"}) {
				t.Fatalf("component=%#v", component)
			}
			completed(t, done, false)
			closed(t, conn)
			if _, err := conn.Write(loginPacket("Player_123", tc.tail)); err == nil {
				t.Fatal("accepted post-disconnect traffic")
			}
			if wakes.Load() != 1 || statuses.Load() != 0 {
				t.Fatalf("calls status=%d join=%d", statuses.Load(), wakes.Load())
			}
		})
	}
}

func TestMalformedRequestsNeverCallCallbacks(t *testing.T) {
	negative := []byte{0xff, 0xff, 0xff, 0xff, 0x0f}
	cases := []struct {
		name string
		wire []byte
	}{
		{"empty", nil}, {"negative length", negative},
		{"overflow length", []byte{0x80, 0x80, 0x80, 0x80, 0x10}},
		{"six byte length", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0}},
		{"overlong length", []byte{0x81, 0}}, {"zero length", []byte{0}},
		{"oversized frame", wireVar(8193)}, {"partial length", []byte{0x80}},
		{"truncated frame", []byte{4, 0, 1}}, {"missing frame", []byte{1}},
		{"unknown handshake ID", wirePacket(1, handshakeBody(771, 2, "host"))},
		{"missing handshake body", wirePacket(0, nil)},
		{"negative protocol", wirePacket(0, negative)},
		{"negative host length", wirePacket(0, append(wireVar(771), negative...))},
		{"long host", wirePacket(0, handshakeBody(771, 2, strings.Repeat("a", 256)))},
		{"long UTF16 host", wirePacket(0, handshakeBody(771, 2, strings.Repeat("😀", 128)))},
		{"bad UTF8 host", wirePacket(0, handshakeBody(771, 2, "\xff"))},
		{"wrong state", handshake(771, 4)}, {"zero state", handshake(771, 0)},
		{"old transfer", handshake(765, 3)}, {"pre1.7 login", handshake(3, 2)},
		{"handshake tail", wirePacket(0, append(handshakeBody(771, 2, "host"), 0))},
		{"wrong status ID", append(handshake(771, 1), wirePacket(1, nil)...)},
		{"status body", append(handshake(771, 1), wirePacket(0, []byte{0})...)},
		{"wrong login ID", append(handshake(771, 2), wirePacket(1, wireString("Player"))...)},
		{"negative login ID", append(handshake(771, 2), append(wireVar(5), negative...)...)},
	}
	for _, name := range []string{"", "abcdefghijklmnopq", "bad name", "bad-name", "pläyer", "\xff", "x\x00y", "x\ny"} {
		cases = append(cases, struct {
			name string
			wire []byte
		}{"username " + fmt.Sprintf("%q", name), append(handshake(771, 2), loginPacket(name, make([]byte, 16))...)})
	}
	for _, tc := range []struct {
		name     string
		protocol int
		tail     []byte
	}{
		{"old extra", 758, []byte{0}}, {"old UUID", 47, make([]byte, 16)},
		{"759 missing flag", 759, nil}, {"759 bool", 759, []byte{2}},
		{"759 timestamp", 759, []byte{1, 0}}, {"759 missing key", 759, append([]byte{1}, make([]byte, 8)...)},
		{"759 long key", 759, append(append([]byte{1}, make([]byte, 8)...), wireVar(8192)...)},
		{"759 negative key", 759, append(append([]byte{1}, make([]byte, 8)...), negative...)},
		{"759 truncated signature", 759, append(append([]byte{1}, make([]byte, 8)...), 0, 5)},
		{"759 unexpected UUID", 759, append([]byte{0}, make([]byte, 16)...)},
		{"760 missing UUID flag", 760, []byte{0}}, {"760 short UUID", 760, []byte{0, 1, 0}},
		{"761 bad UUID flag", 761, []byte{2}}, {"761 short UUID", 761, []byte{1, 0}},
		{"763 missing UUID flag", 763, nil}, {"763 extra", 763, []byte{0, 0}},
		{"764 missing UUID", 764, nil}, {"771 short UUID", 771, make([]byte, 15)},
		{"771 old optional UUID", 771, append([]byte{1}, make([]byte, 16)...)},
		{"771 extra", 771, make([]byte, 17)},
	} {
		cases = append(cases, struct {
			name string
			wire []byte
		}{tc.name, append(handshake(tc.protocol, 2), loginPacket("Player", tc.tail)...)})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			conn, done := start(t, func(int32) Status { calls.Add(1); return Status{} }, func(Join) string { calls.Add(1); return "" })
			// Some inputs are rejected before the writer finishes, which is expected.
			if len(tc.wire) != 0 {
				_, _ = conn.Write(tc.wire)
			}
			conn.Close()
			completed(t, done, true)
			if calls.Load() != 0 {
				t.Fatalf("malformed input called callback %d times", calls.Load())
			}
		})
	}
}

func TestOutputBounds(t *testing.T) {
	for _, motd := range []string{strings.Repeat("a", maxStatusBytes+1), strings.Repeat("\x00", maxStatusBytes/2)} {
		conn, done := start(t, func(int32) Status { return Status{MOTD: motd} }, nil)
		send(t, conn, append(handshake(771, 1), wirePacket(0, nil)...))
		closed(t, conn)
		completed(t, done, true)
	}
	for _, message := range []string{strings.Repeat("a", maxMessageBytes+1), strings.Repeat("\x00", maxMessageBytes)} {
		conn, done := start(t, nil, func(Join) string { return message })
		send(t, conn, append(handshake(771, 2), loginPacket("Player", make([]byte, 16))...))
		encoded := receiveJSON(t, conn)
		if len(encoded) > maxMessageBytes {
			t.Fatalf("disconnect length=%d", len(encoded))
		}
		var component textComponent
		if err := json.Unmarshal(encoded, &component); err != nil {
			t.Fatal(err)
		}
		if component.Text != fallbackMessage {
			t.Fatalf("message=%q", component.Text)
		}
		completed(t, done, false)
	}
}

type recordingConn struct {
	net.Conn
	deadlines []time.Time
}

func (c *recordingConn) SetDeadline(d time.Time) error {
	c.deadlines = append(c.deadlines, d)
	return c.Conn.SetDeadline(d)
}
func TestFixedConnectionDeadline(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	conn := &recordingConn{Conn: server}
	started := time.Now()
	done := make(chan error, 1)
	go func() { done <- Handle(conn, nil, nil) }()
	// Make progress late in the handshake. A per-packet deadline reset would
	// cause this test to take longer or record another SetDeadline call.
	time.Sleep(150 * time.Millisecond)
	if _, err := client.Write(handshake(771, 2)); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		var timeout net.Error
		if !errors.As(err, &timeout) || !timeout.Timeout() {
			t.Fatalf("expected timeout, got %v", err)
		}
	case <-time.After(7 * time.Second):
		t.Fatal("deadline not enforced")
	}
	elapsed := time.Since(started)
	if elapsed < 4500*time.Millisecond || elapsed > 7*time.Second {
		t.Fatalf("deadline elapsed=%s", elapsed)
	}
	if len(conn.deadlines) != 1 {
		t.Fatalf("deadline set %d times", len(conn.deadlines))
	}
	if delta := conn.deadlines[0].Sub(started); delta < 5*time.Second || delta > 5500*time.Millisecond {
		t.Fatalf("deadline delta=%s", delta)
	}
	closed(t, client)
}

func TestPipelinedLoginWakesOnceAndDoesNotForward(t *testing.T) {
	var calls atomic.Int32
	conn, done := start(t, nil, func(Join) string { calls.Add(1); return "Starting" })
	packet := loginPacket("Player", make([]byte, 16))
	wire := append(append(handshake(771, 2), packet...), packet...)
	written := make(chan int, 1)
	go func() { n, _ := conn.Write(wire); written <- n }()
	receiveJSON(t, conn)
	completed(t, done, false)
	closed(t, conn)
	if n := <-written; n != len(wire)-len(packet) {
		t.Fatalf("consumed post-login traffic: %d bytes", n)
	}
	if calls.Load() != 1 {
		t.Fatalf("join calls=%d", calls.Load())
	}
}

func TestJoinStillCalledOnceIfDisconnectWriteFails(t *testing.T) {
	var calls atomic.Int32
	conn, done := start(t, nil, func(Join) string { calls.Add(1); return "Starting" })
	send(t, conn, append(handshake(771, 2), loginPacket("Player", make([]byte, 16))...))
	conn.Close()
	completed(t, done, true)
	if calls.Load() != 1 {
		t.Fatalf("join calls=%d", calls.Load())
	}
}

func TestVarIntBounds(t *testing.T) {
	for _, n := range []uint64{0, 1, 127, 128, 255, 8192, 16384, 1 << 28, 1<<31 - 1} {
		actual, err := readVarInt(bytes.NewReader(wireVar(n)))
		if err != nil || actual != int32(n) {
			t.Fatalf("VarInt %d = %d, %v", n, actual, err)
		}
	}
}

func TestWriteDeadline(t *testing.T) {
	var wakes atomic.Int32
	conn, done := start(t, func(int32) Status { return Status{MOTD: "Offline"} }, func(Join) string {
		wakes.Add(1)
		return ""
	})
	send(t, conn, append(handshake(771, 1), wirePacket(0, nil)...))
	// Do not read: net.Pipe blocks the status write until the same deadline.
	select {
	case err := <-done:
		var timeout net.Error
		if !errors.As(err, &timeout) || !timeout.Timeout() {
			t.Fatalf("expected write timeout, got %v", err)
		}
	case <-time.After(7 * time.Second):
		t.Fatal("write deadline not enforced")
	}
	closed(t, conn)
	if wakes.Load() != 0 {
		t.Fatal("blocked status write woke server")
	}
}

func TestEveryLoginTruncation(t *testing.T) {
	for _, protocol := range []int{47, 760, 771} {
		var tail []byte
		if protocol == 760 {
			tail = append(append(signatureTail(), 1), make([]byte, 16)...)
		} else if protocol == 771 {
			tail = make([]byte, 16)
		}
		wire := append(handshake(protocol, 2), loginPacket("Player", tail)...)
		for end := 0; end < len(wire); end++ {
			t.Run(fmt.Sprintf("protocol%d/prefix%d", protocol, end), func(t *testing.T) {
				var wakes atomic.Int32
				conn, done := start(t, nil, func(Join) string { wakes.Add(1); return "" })
				if end != 0 {
					send(t, conn, wire[:end])
				}
				conn.Close()
				completed(t, done, true)
				if wakes.Load() != 0 {
					t.Fatal("truncated login woke server")
				}
			})
		}
	}
}
