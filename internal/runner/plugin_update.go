package runner

import (
	"context"
	"errors"
	"sync"
)

// BeginPluginUpdateOperation reuses the stopped-process/run-lock/exclusive-root
// fence and joins daemon shutdown like managed imports. The watcher is joined,
// never detached; the caller must finish publication/rollback before release.
func (m *ServerManager) BeginPluginUpdateOperation(ctx context.Context, id string) (context.Context, func(), error) {
	release, err := m.BeginProxyConfigOperation(ctx, id)
	if err != nil {
		return nil, nil, errors.New("stop the server and wait for active file or power operations before updating plugins")
	}
	m.launchMu.Lock()
	if m.isShuttingDown() {
		m.launchMu.Unlock()
		release()
		return nil, nil, errDaemonShuttingDown
	}
	m.importWG.Add(1)
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-m.shutdownCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	m.launchMu.Unlock()
	var once sync.Once
	return ctx, func() { once.Do(func() { cancel(); <-done; release(); m.importWG.Done() }) }, nil
}

// BlockPluginUpdateRecovery uses the same fail-closed power/recovery gate as a
// failed restore. Only operator repair can resolve a failed publication rollback.
func (m *ServerManager) BlockPluginUpdateRecovery(id string) {
	m.restoreBlocked.Store(id, errors.New("plugin update rollback failed; restore the retained plugin backup before operator recovery"))
}
