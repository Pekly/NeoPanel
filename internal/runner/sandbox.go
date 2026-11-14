package runner

import (
	"fmt"
	"path/filepath"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
)

// Mount the resolver file, never its parent (/etc on many hosts).
func sandboxResolverArgs(path string) []string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil
	}
	return []string{"--ro-bind-try", resolved, "/etc/resolv.conf"}
}

func systemdLimitsLabel(srv database.Server, tasksMax int) string {
	label := fmt.Sprintf("TasksMax=%d", tasksMax)
	if downloader.IsNativeSoftware(srv.Type) && srv.MemoryMaxMB > 0 {
		label += fmt.Sprintf("; MemoryMax=%d MiB enforced by systemd", srv.MemoryMaxMB)
	}
	return label
}
