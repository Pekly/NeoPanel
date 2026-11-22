package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

var ErrProxyConfigRunning = errors.New("stop the proxy before changing its configuration")

// BeginProxyConfigOperation excludes power, restore and import operations without
// stopping a running proxy. The caller must release the gate after all file work.
func (m *ServerManager) BeginProxyConfigOperation(ctx context.Context, id string) (func(), error) {
	release, err := m.beginOperation(id)
	if err != nil {
		return nil, err
	}
	releaseFiles, err := rooted.TryExclusiveMutation(m.GetWorkingDir(id))
	if err != nil {
		release()
		return nil, fmt.Errorf("%w: %v", ErrServerOperation, err)
	}
	releaseOperation := release
	release = func() { releaseFiles(); releaseOperation() }
	m.launchMu.Lock()
	switch {
	case m.isShuttingDown():
		err = errDaemonShuttingDown
	case ctx.Err() != nil:
		err = ctx.Err()
	case m.GetStatus(id) != "offline":
		err = ErrProxyConfigRunning
	default:
		err = m.invalidateIdleWake(id)
		if err == nil {
			m.mu.Lock()
			m.cancelRetryLocked(id)
			m.restorePaused.Store(id, true)
			m.mu.Unlock()
		}
	}
	m.launchMu.Unlock()
	if err != nil {
		release()
		return nil, err
	}
	lock, err := acquireRunLock(m.GetWorkingDir(id))
	if err != nil {
		release()
		return nil, fmt.Errorf("proxy configuration is busy: %w", err)
	}
	if err := ensureServerProcessesStopped(m.GetWorkingDir(id)); err != nil {
		lock.Close()
		release()
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		lock.Close()
		release()
		return nil, err
	}
	var once sync.Once
	return func() { once.Do(func() { lock.Close(); release() }) }, nil
}
