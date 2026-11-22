package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type catalogTransport func(*http.Request) (*http.Response, error)

func (f catalogTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestVersionCatalogRefreshCoalescesAndCoolsDown(t *testing.T) {
	versionsMu.Lock()
	old, expiry, refresh := versionsCached, versionsExpiresAt, versionsRefreshAfter
	versionsCached, versionsExpiresAt, versionsRefreshAfter = nil, time.Time{}, time.Time{}
	versionsMu.Unlock()
	transport := http.DefaultTransport
	var calls atomic.Int32
	gate := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(gate) }) }
	defer unblock()
	http.DefaultTransport = catalogTransport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		select {
		case <-gate:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("unavailable")), Header: make(http.Header)}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = transport
		versionsMu.Lock()
		versionsCached, versionsExpiresAt, versionsRefreshAfter = old, expiry, refresh
		versionsMu.Unlock()
	})
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := cachedVersionsCatalog(context.Background(), true); err != nil {
				t.Error(err)
			}
		}()
	}
	deadline := time.Now().Add(3 * time.Second)
	for calls.Load() < 8 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := cachedVersionsCatalog(ctx, true); err != context.Canceled {
		t.Error("cancelled waiter not released", err)
	}
	unblock()
	wg.Wait()
	// Some providers have one sequential fallback, so record this single batch
	// rather than assuming every provider needs exactly one HTTP request.
	count := calls.Load()
	if count < 8 || count > 16 {
		t.Fatalf("amplified fetch batch: %d requests", count)
	}
	for i := 0; i < 20; i++ {
		if _, err := cachedVersionsCatalog(context.Background(), true); err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != count {
		t.Fatal("forced refresh bypassed cooldown")
	}
}

func TestVersionProviderJSONBounded(t *testing.T) {
	for _, body := range []string{`{} {}`, `{"value":"` + strings.Repeat("x", 4<<20) + `"}`} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) }))
		var out any
		ok := versionGetJSON(context.Background(), srv.URL, &out)
		srv.Close()
		if ok {
			t.Fatal("accepted unbounded or concatenated metadata")
		}
	}
}
