package runner

import (
	"errors"
	"fmt"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

// beginFileOperation keeps managed file handles from publishing into a tree
// moved to recovery or removed by restore/import/reinstall/deletion.
func (m *ServerManager) beginFileOperation(id string) (func(), error) {
	unlock, err := m.beginOperation(id)
	if err != nil {
		return nil, err
	}
	releaseFiles, err := rooted.TryExclusiveMutation(m.GetWorkingDir(id))
	if err != nil {
		unlock()
		if errors.Is(err, rooted.ErrMutationBusy) {
			return nil, fmt.Errorf("%w: managed file writes are active", ErrServerOperation)
		}
		return nil, err
	}
	return func() { releaseFiles(); unlock() }, nil
}
