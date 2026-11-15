//go:build !windows

package runner

import (
	"fmt"
	"runtime"
)

func (pm *ProcessMetrics) readPlatformStatsLocked() (StatsSnapshot, error) {
	if runtime.GOOS == "linux" {
		return pm.readLinuxStatsLocked()
	}
	return StatsSnapshot{}, fmt.Errorf("process metrics unsupported on %s", runtime.GOOS)
}
