//go:build windows

package runner

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var getProcessMemoryInfo = windows.NewLazySystemDLL("kernel32.dll").NewProc("K32GetProcessMemoryInfo")

type processMemoryCounters struct {
	Size, PageFaultCount                               uint32
	PeakWorkingSetSize, WorkingSetSize                 uintptr
	QuotaPeakPagedPoolUsage, QuotaPagedPoolUsage       uintptr
	QuotaPeakNonPagedPoolUsage, QuotaNonPagedPoolUsage uintptr
	PagefileUsage, PeakPagefileUsage                   uintptr
}

func fileTimeTicks(t windows.Filetime) uint64 {
	return uint64(t.HighDateTime)<<32 | uint64(t.LowDateTime)
}

// Root-process metrics, not aggregate job metrics. Working set is resident RAM;
// CPU is normalized to total host logical CPUs, matching Linux's convention.
// Windows has no per-process network byte counters here: report zero/unsupported.
func (pm *ProcessMetrics) readPlatformStatsLocked() (StatsSnapshot, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ|windows.SYNCHRONIZE, false, uint32(pm.PID))
	if err != nil {
		return StatsSnapshot{}, err
	}
	defer windows.CloseHandle(h)
	if state, err := windows.WaitForSingleObject(h, 0); err != nil || state != uint32(windows.WAIT_TIMEOUT) {
		return StatsSnapshot{}, fmt.Errorf("process is not running (wait=%d): %v", state, err)
	}
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &created, &exited, &kernel, &user); err != nil {
		return StatsSnapshot{}, err
	}
	identity := fileTimeTicks(created)
	if pm.processCreated != 0 && pm.processCreated != identity {
		return StatsSnapshot{}, fmt.Errorf("process identity changed")
	}
	var memory processMemoryCounters
	memory.Size = uint32(unsafe.Sizeof(memory))
	ok, _, callErr := getProcessMemoryInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&memory)), uintptr(memory.Size))
	if ok == 0 {
		return StatsSnapshot{}, fmt.Errorf("read process working set: %w", callErr)
	}
	now := time.Now()
	total := int64(fileTimeTicks(kernel) + fileTimeTicks(user))
	if pm.processCreated != 0 {
		elapsed := now.Sub(pm.lastReadTime).Seconds()
		if elapsed >= 0.2 {
			delta := total - pm.lastCPUTime
			if delta < 0 {
				delta = 0
			}
			pm.CPUPercent = float64(delta) / 1e7 / elapsed * 100 / float64(runtime.NumCPU())
			if pm.CPUPercent > 100 {
				pm.CPUPercent = 100
			}
			pm.lastCPUTime, pm.lastReadTime = total, now
		}
	} else {
		pm.lastCPUTime, pm.lastReadTime = total, now
	}
	pm.processCreated = identity
	pm.MemoryMB = int64(memory.WorkingSetSize / (1024 * 1024))
	return StatsSnapshot{CPUPercent: pm.CPUPercent, MemoryMB: pm.MemoryMB}, nil
}
