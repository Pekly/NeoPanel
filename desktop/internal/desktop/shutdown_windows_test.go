//go:build windows && !node

package desktop

import (
	"context"
	"errors"
	"testing"
	"time"
)

type shutdownProbe struct{ call func(context.Context) error }

func (p shutdownProbe) Shutdown(ctx context.Context) error { return p.call(ctx) }

func TestDesktopShutdownBoundsHTTPDrainAndWaitsForCleanup(t *testing.T) {
	entered, finish, done := make(chan context.Context, 1), make(chan struct{}), make(chan error, 1)
	want := errors.New("cleanup result")
	go func() {
		done <- shutdownBackend(shutdownProbe{call: func(ctx context.Context) error {
			entered <- ctx
			<-finish
			return want
		}})
	}()
	ctx := <-entered
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 15*time.Second {
		t.Fatal("desktop HTTP drain is not bounded")
	}
	select {
	case <-done:
		t.Fatal("returned before cleanup completed")
	default:
	}
	close(finish)
	if err := <-done; !errors.Is(err, want) {
		t.Fatal(err)
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatal("shutdown context leaked")
	}
}
