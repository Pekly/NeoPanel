//go:build windows && !node

package desktop

import (
	"context"
	"net"
	"testing"

	"github.com/Pekly/NeoPanel/internal/config"
)

func TestSingleInstanceLock(t *testing.T) {
	id := InstanceID(t.TempDir())
	release, acquired, err := AcquireInstance(id)
	if err != nil || !acquired {
		t.Fatal(acquired, err)
	}
	_, second, err := AcquireInstance(id)
	if err != nil || second {
		t.Fatal(second, err)
	}
	release()
	release, acquired, err = AcquireInstance(id)
	if err != nil || !acquired {
		t.Fatal(acquired, err)
	}
	release()
}

func TestDesktopRecoversOccupiedPortAndPersistsOrigin(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	cfg.Server.Port = occupied.Addr().(*net.TCPAddr).Port
	svc, err := startBackend(cfg)
	if err != nil {
		t.Fatal(err)
	}
	firstURL := svc.URL()
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	saved, err := config.LoadConfig(cfg.FilePath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Server.Port == 0 || saved.Server.Port == occupied.Addr().(*net.TCPAddr).Port {
		t.Fatal("did not persist recovered port")
	}
	svc, err = startBackend(saved)
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown(context.Background())
	if svc.URL() != firstURL {
		t.Fatalf("origin changed on restart: %s -> %s", firstURL, svc.URL())
	}
}
