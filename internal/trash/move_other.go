//go:build !linux && !windows

package trash

import (
	"fmt"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

func moveNoReplace(from *rooted.Root, old string, to *rooted.Root, name string) error {
	return fmt.Errorf("atomic trash moves are supported only on Linux and Windows")
}
