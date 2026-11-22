package sftp

import (
	"io"
	"net"
	"testing"
	"time"

	sftpclient "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func waitAdmission(t *testing.T, s *SFTPServer, total, pending int) {
	t.Helper()
	until := time.Now().Add(3 * time.Second)
	for {
		s.admission.mu.Lock()
		gotTotal, gotPending := len(s.admission.connections), s.admission.pending
		s.admission.mu.Unlock()
		if gotTotal == total && gotPending == pending {
			return
		}
		if time.Now().After(until) {
			t.Fatalf("admission total/pending = %d/%d, want %d/%d", gotTotal, gotPending, total, pending)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func silentPeer(t *testing.T, f *fixture) net.Conn {
	t.Helper()
	c, err := net.DialTimeout("tcp", f.addr, time.Second)
	must(t, err)
	t.Cleanup(func() { c.Close() })
	return c
}

func expectPeerClosed(t *testing.T, c net.Conn) {
	t.Helper()
	must(t, c.SetReadDeadline(time.Now().Add(2*time.Second)))
	_, err := io.Copy(io.Discard, c) // consume the server identification, if any
	if e, ok := err.(net.Error); ok && e.Timeout() {
		t.Fatal("peer was not closed by server")
	}
}

func TestSilentHandshakeTimeoutReleasesAdmission(t *testing.T) {
	f := newFixture(t)
	f.server.admission.mu.Lock()
	f.server.admission.handshakeTimeout = 100 * time.Millisecond
	f.server.admission.mu.Unlock()
	c := silentPeer(t, f)
	waitAdmission(t, f.server, 1, 1)
	expectPeerClosed(t, c)
	waitAdmission(t, f.server, 0, 0)
	// Successful auth clears the same injected deadline, not just the slot.
	f.server.admission.mu.Lock()
	f.server.admission.handshakeTimeout = 500 * time.Millisecond
	f.server.admission.mu.Unlock()
	_, client := f.connect(t, "owner", "key")
	waitAdmission(t, f.server, 1, 0)
	time.Sleep(600 * time.Millisecond)
	_, err := client.Stat("/file.txt")
	must(t, err)
}

func TestPendingAdmissionCapsRejectAndRelease(t *testing.T) {
	for _, kind := range []string{"global", "source"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			f.server.admission.mu.Lock()
			if kind == "global" {
				f.server.admission.maxPending = 1
			} else {
				f.server.admission.maxPerSource = 1
			}
			f.server.admission.mu.Unlock()
			first := silentPeer(t, f)
			waitAdmission(t, f.server, 1, 1)
			rejected := silentPeer(t, f)
			expectPeerClosed(t, rejected)
			waitAdmission(t, f.server, 1, 1)
			must(t, first.Close())
			waitAdmission(t, f.server, 0, 0)
			_, client := f.connect(t, "owner", "password")
			_, err := client.Stat("/file.txt")
			must(t, err)
			waitAdmission(t, f.server, 1, 0)
		})
	}
}

func TestAuthenticatedConnectionAndChannelCaps(t *testing.T) {
	f := newFixture(t)
	f.server.admission.mu.Lock()
	f.server.admission.maxTotal = 1
	f.server.admission.mu.Unlock()
	connection, client := f.connect(t, "owner", "key")
	waitAdmission(t, f.server, 1, 0)
	rejected := silentPeer(t, f)
	expectPeerClosed(t, rejected)
	_, err := client.Stat("/file.txt")
	must(t, err)
	channels := make([]ssh.Channel, 0, maxSessionChannels-1)
	for i := 1; i < maxSessionChannels; i++ {
		ch, reqs, err := connection.OpenChannel("session", nil)
		must(t, err)
		go ssh.DiscardRequests(reqs)
		channels = append(channels, ch)
	}
	_, _, err = connection.OpenChannel("session", nil)
	denied(t, err)
	_, err = client.Stat("/file.txt")
	must(t, err)
	for _, ch := range channels {
		must(t, ch.Close())
	}
	// Channel close is asynchronous; retry until one slot becomes available.
	until := time.Now().Add(2 * time.Second)
	for {
		c, err := sftpclient.NewClient(connection)
		if err == nil {
			must(t, c.Close())
			break
		}
		if time.Now().After(until) {
			t.Fatal("channel slots not released:", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
	must(t, connection.Close())
	waitAdmission(t, f.server, 0, 0)
	_, replacement := f.connect(t, "owner", "key")
	_, err = replacement.Stat("/file.txt")
	must(t, err)
}

func TestShutdownClosesPendingAndAuthenticatedConnections(t *testing.T) {
	f := newFixture(t)
	_, _ = f.connect(t, "owner", "key")
	pending := silentPeer(t, f)
	waitAdmission(t, f.server, 2, 1)
	done := make(chan struct{})
	go func() { f.server.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop did not drain connection workers")
	}
	waitAdmission(t, f.server, 0, 0)
	expectPeerClosed(t, pending)
	f.server.Stop() // idempotent, including fixture cleanup
}

func TestFailedAuthenticationReleasesAdmission(t *testing.T) {
	f := newFixture(t)
	c, err := f.dial("outsider#srv-alpha", "password")
	if c != nil {
		c.Close()
	}
	denied(t, err)
	waitAdmission(t, f.server, 0, 0)
}
