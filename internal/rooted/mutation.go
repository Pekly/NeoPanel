package rooted

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ErrMutationBusy means a managed file mutation or exclusive maintenance job is
// active. Callers must retry, never wait while holding manager/launch gates.
var ErrMutationBusy = errors.New("server files are busy; retry after the active operation finishes")

var mutations = struct {
	sync.Mutex
	roots map[string]*mutationState
}{roots: make(map[string]*mutationState)}

type mutationState struct {
	readers   int
	exclusive bool
}

// TryMutation leases an ordinary managed mutation through its final write/close.
// Reads need no lease; ordinary edits (including running-server edits) coexist.
func TryMutation(dir string) (func(), error) { return tryMutation(dir, false) }

// TryExclusiveMutation excludes managed API/SFTP writes for a complete job,
// including snapshot, publication and rollback. Acquire after the manager operation
// gate and release before that gate. These process-local leases cannot coordinate
// unsupported external host edits or other daemons writing the same tree.
func TryExclusiveMutation(dir string) (func(), error) { return tryMutation(dir, true) }

func tryMutation(dir string, exclusive bool) (func(), error) {
	key, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	key, err = filepath.EvalSymlinks(key)
	if err != nil {
		return nil, err
	}
	key = filepath.Clean(key)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	mutations.Lock()
	defer mutations.Unlock()
	s := mutations.roots[key]
	if s == nil {
		s = &mutationState{}
		mutations.roots[key] = s
	}
	if s.exclusive || (exclusive && s.readers != 0) {
		return nil, ErrMutationBusy
	}
	if exclusive {
		s.exclusive = true
	} else {
		s.readers++
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			mutations.Lock()
			defer mutations.Unlock()
			if exclusive {
				s.exclusive = false
			} else {
				s.readers--
			}
			if !s.exclusive && s.readers == 0 {
				delete(mutations.roots, key)
			}
		})
	}, nil
}
