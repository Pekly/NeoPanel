package runner

import (
	"fmt"
	"os"
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

// OSHI and other runtime libraries expect /etc/os-release, which may be a
// symlink to /usr/lib/os-release. Expose only this public regular file.
func sandboxOSReleaseArgs(candidates ...string) []string {
	for _, candidate := range candidates {
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			continue
		}
		info, err := os.Stat(resolved)
		if err == nil && info.Mode().IsRegular() {
			return []string{"--ro-bind-try", resolved, "/etc/os-release"}
		}
	}
	return nil
}

func systemdLimitsLabel(srv database.Server, tasksMax int) string {
	label := fmt.Sprintf("TasksMax=%d", tasksMax)
	if memoryMax := serverMemoryCeilingMB(srv); memoryMax > 0 {
		label += fmt.Sprintf("; MemoryMax=%d MiB enforced by systemd", memoryMax)
	}
	return label
}

// MemoryMaxMB is a Java HEAP budget, not the whole process budget. Reserve
// max(512 MiB, 50% of heap) for metaspace, code cache, thread stacks, direct
// buffers and other native allocations. This is a bounded allowance, not a
// guarantee for every plugin/modpack. Native servers retain their exact limit.
// A systemd cgroup is required; bwrap alone and process/none do not enforce it.
func serverMemoryCeilingMB(srv database.Server) int64 {
	if srv.MemoryMaxMB <= 0 {
		return 0
	}
	heap := int64(srv.MemoryMaxMB)
	if downloader.IsNativeSoftware(srv.Type) {
		return heap
	}
	overhead := heap / 2
	if overhead < 512 {
		overhead = 512
	}
	// Saturate rather than wrapping a positive configured limit to unlimited.
	const maxInt64 = int64(1<<63 - 1)
	if heap > maxInt64-overhead {
		return maxInt64
	}
	return heap + overhead
}
