package mcwake

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
	"unicode/utf8"
)

const (
	connectionTimeout = 5 * time.Second
	maxInputFrame     = 8192
	maxStatusBytes    = 128 * 1024
	maxMessageBytes   = 4096
	fallbackMessage   = "Server is starting. Please reconnect shortly."
)

// Status describes a stopped server. IconDataURL is optional and must already
// be a validated PNG data URI. MaxPlayers and text are supplied by the caller.
type Status struct {
	MOTD        string
	VersionName string
	MaxPlayers  int
	IconDataURL string
}

// Join is an unauthenticated, structurally valid login request.
type Join struct {
	Protocol int32
	Username string
}

type textComponent struct {
	Text string `json:"text"`
}

// Handle serves one status exchange or login attempt and always closes conn.
// status is called only for a valid status request; join is called once only
// after a valid Login Start. join's result is the login disconnect text.
// Only the callback for the requested state is required to be non-nil.
// A clean EOF after a status response is successful; malformed or truncated
// input, timeouts and I/O failures return errors without exposing input text.
func Handle(conn net.Conn, status func(protocol int32) Status, join func(Join) string) error {
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(connectionTimeout)); err != nil {
		return err
	}
	id, body, err := readPacket(conn)
	if err != nil {
		return err
	}
	if id != 0 {
		return errors.New("mcwake: expected handshake")
	}
	protocol, err := readVarInt(body)
	if err != nil {
		return err
	}
	if _, err = readString(body, 255); err != nil {
		return err
	}
	if err = skip(body, 2); err != nil { // Port is deliberately unused.
		return err
	}
	state, err := readVarInt(body)
	if err != nil {
		return err
	}
	if body.Len() != 0 {
		return errors.New("mcwake: trailing handshake data")
	}
	switch state {
	case 1:
		return handleStatus(conn, protocol, status)
	case 2, 3:
		if protocol < 4 || (state == 3 && protocol < 766) {
			return errors.New("mcwake: unsupported login protocol or state")
		}
		return handleLogin(conn, protocol, join)
	default:
		return errors.New("mcwake: unsupported handshake state")
	}
}

func handleStatus(conn net.Conn, protocol int32, status func(int32) Status) error {
	id, body, err := readPacket(conn)
	if err != nil {
		return err
	}
	if id != 0 || body.Len() != 0 {
		return errors.New("mcwake: invalid status request")
	}
	if status == nil {
		return errors.New("mcwake: missing status callback")
	}
	s := status(protocol)
	// Preflight each string to avoid marshaling unbounded caller data.
	if len(s.MOTD) > maxStatusBytes || len(s.VersionName) > maxStatusBytes || len(s.IconDataURL) > maxStatusBytes {
		return errors.New("mcwake: status response too large")
	}
	response := struct {
		Version struct {
			Name     string `json:"name"`
			Protocol int32  `json:"protocol"`
		} `json:"version"`
		Players struct {
			Max    int `json:"max"`
			Online int `json:"online"`
		} `json:"players"`
		Description textComponent `json:"description"`
		Favicon     string        `json:"favicon,omitempty"`
	}{Description: textComponent{s.MOTD}, Favicon: s.IconDataURL}
	response.Version.Name = s.VersionName
	response.Version.Protocol = protocol
	response.Players.Max = s.MaxPlayers
	encoded, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if err = writePacket(conn, 0, stringBody(encoded), maxStatusBytes); err != nil {
		return err
	}
	id, body, err = readPacket(conn)
	if err == io.EOF { // Closing without a ping is allowed, truncating one is not.
		return nil
	}
	if err != nil {
		return err
	}
	if id != 1 || body.Len() != 8 {
		return errors.New("mcwake: invalid status ping")
	}
	var timestamp [8]byte
	if _, err = io.ReadFull(body, timestamp[:]); err != nil {
		return err
	}
	return writePacket(conn, 1, timestamp[:], maxStatusBytes)
}

func handleLogin(conn net.Conn, protocol int32, join func(Join) string) error {
	id, body, err := readPacket(conn)
	if err != nil {
		return err
	}
	if id != 0 {
		return errors.New("mcwake: expected Login Start")
	}
	username, err := readString(body, 16)
	if err != nil {
		return err
	}
	if len(username) == 0 || len(username) > 16 {
		return errors.New("mcwake: invalid username")
	}
	for _, c := range []byte(username) {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return errors.New("mcwake: invalid username")
		}
	}
	if err = loginTail(body, protocol); err != nil {
		return err
	}
	if body.Len() != 0 {
		return errors.New("mcwake: trailing Login Start data")
	}
	if join == nil {
		return errors.New("mcwake: missing join callback")
	}
	message := join(Join{Protocol: protocol, Username: username})
	if len(message) > maxMessageBytes {
		message = fallbackMessage
	}
	encoded, err := json.Marshal(textComponent{message})
	if err != nil {
		return err
	}
	if len(encoded) > maxMessageBytes {
		encoded, _ = json.Marshal(textComponent{fallbackMessage})
	}
	return writePacket(conn, 0, stringBody(encoded), maxMessageBytes+16)
}

func loginTail(body *bytes.Reader, protocol int32) error {
	if protocol == 759 || protocol == 760 {
		present, err := readBool(body)
		if err != nil {
			return err
		}
		if present {
			if err = skip(body, 8); err != nil { // Signature timestamp.
				return err
			}
			for i := 0; i < 2; i++ { // Public key and signature, bounded by frame.
				n, err := readVarInt(body)
				if err != nil {
					return err
				}
				if err = skip(body, int(n)); err != nil {
					return err
				}
			}
		}
	}
	if protocol >= 760 && protocol <= 763 {
		present, err := readBool(body)
		if err != nil {
			return err
		}
		if present {
			return skip(body, 16)
		}
	} else if protocol >= 764 {
		return skip(body, 16)
	}
	return nil
}

func readBool(r *bytes.Reader) (bool, error) {
	b, err := r.ReadByte()
	if err != nil {
		return false, io.ErrUnexpectedEOF
	}
	if b > 1 {
		return false, errors.New("mcwake: invalid boolean")
	}
	return b == 1, nil
}

func skip(r *bytes.Reader, n int) error {
	if n < 0 || n > r.Len() {
		return io.ErrUnexpectedEOF
	}
	_, err := r.Seek(int64(n), io.SeekCurrent)
	return err
}

func readString(r *bytes.Reader, maxUnits int) (string, error) {
	n, err := readVarInt(r)
	if err != nil {
		return "", err
	}
	if n > int32(maxUnits*4) {
		return "", errors.New("mcwake: string too long")
	}
	if int(n) > r.Len() {
		return "", io.ErrUnexpectedEOF
	}
	b := make([]byte, int(n))
	if _, err = io.ReadFull(r, b); err != nil {
		return "", err
	}
	if !utf8.Valid(b) {
		return "", errors.New("mcwake: invalid UTF-8")
	}
	units := 0
	for _, c := range string(b) {
		units++
		if c > 0xffff {
			units++
		}
	}
	if units > maxUnits {
		return "", errors.New("mcwake: string too long")
	}
	return string(b), nil
}

// readVarInt rejects negatives, overflow, >5-byte encodings and overlong forms.
// EOF is returned only before the first byte, so partial frames stay errors.
func readVarInt(r io.Reader) (int32, error) {
	var value uint32
	var b [1]byte
	for i := 0; i < 5; i++ {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			if err == io.EOF && i > 0 {
				err = io.ErrUnexpectedEOF
			}
			return 0, err
		}
		if i == 4 && b[0] > 7 {
			return 0, errors.New("mcwake: negative or oversized VarInt")
		}
		value |= uint32(b[0]&0x7f) << (7 * i)
		if b[0]&0x80 == 0 {
			if i > 0 && b[0] == 0 {
				return 0, errors.New("mcwake: noncanonical VarInt")
			}
			return int32(value), nil
		}
	}
	return 0, errors.New("mcwake: oversized VarInt")
}

func readPacket(r io.Reader) (int32, *bytes.Reader, error) {
	n, err := readVarInt(r)
	if err != nil {
		return 0, nil, err
	}
	if n < 1 || n > maxInputFrame {
		return 0, nil, errors.New("mcwake: invalid frame size")
	}
	frame := make([]byte, int(n))
	if _, err = io.ReadFull(r, frame); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	body := bytes.NewReader(frame)
	id, err := readVarInt(body)
	if err != nil {
		return 0, nil, fmt.Errorf("mcwake: invalid packet ID: %w", err)
	}
	return id, body, nil
}

func appendVarInt(dst []byte, n int) []byte {
	for n >= 128 {
		dst = append(dst, byte(n&0x7f)|0x80)
		n >>= 7
	}
	return append(dst, byte(n))
}

func stringBody(s []byte) []byte {
	return append(appendVarInt(nil, len(s)), s...)
}

func writePacket(w io.Writer, id int, body []byte, limit int) error {
	payload := append(appendVarInt(nil, id), body...)
	frame := append(appendVarInt(nil, len(payload)), payload...)
	if len(frame) > limit {
		return errors.New("mcwake: response too large")
	}
	for len(frame) != 0 {
		n, err := w.Write(frame)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		frame = frame[n:]
	}
	return nil
}
