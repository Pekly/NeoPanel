//go:build linux

package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Pekly/NeoPanel/internal/chunker"
	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/sys/unix"
)

func conversionPlatform(srv *database.Server) error {
	if srv.MemoryMaxMB < 512 || srv.MemoryMaxMB > 1<<20 || srv.CPULimitPercent <= 0 || srv.CPULimitPercent > 100000 {
		return errors.New("conversion requires an assigned memory limit of 512 MiB or more and a positive CPU quota")
	}
	for _, tool := range []string{"systemd-run", "systemctl", "bwrap", "/bin/sh", "/usr/bin/tar", "/bin/cp"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("world conversion requires Linux systemd+cgroup v2 and Bubblewrap: missing %s", tool)
		}
	}
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		return errors.New("conversion requires unified cgroup v2 resource enforcement")
	}
	return nil
}

// The fixed bootstrap checks effective kernel limits BEFORE executing Bubblewrap
// or Java. Tool presence, a JVM heap flag and systemd exit 0 are not sufficient.
const conversionQuotaCheck = `set -eu
cg=$(sed -n 's/^0:://p' /proc/self/cgroup)
case "$cg" in /*) ;; *) echo 'Cannot verify cgroup' >&2; exit 71;; esac
mem=$(cat "/sys/fs/cgroup$cg/memory.max")
case "$mem" in ''|*[!0-9]*) echo 'Memory controller is not enforcing a ceiling' >&2; exit 72;; esac
[ "$mem" -le "$1" ] || exit 72
read quota period < "/sys/fs/cgroup$cg/cpu.max"
case "$quota:$period" in *max*|*[!0-9:]*) echo 'CPU controller is not enforcing a quota' >&2; exit 73;; esac
[ "$period" -gt 0 ] && [ $((quota * 100)) -le $(($2 * period)) ] || exit 73
shift 2
exec "$@"
`
const conversionSandboxScript = `set -eu
mkdir /work/input /work/output
cp -R /source/. /work/input/
"$@" >&2
exec /usr/bin/tar -C /work/output -cf - .
`

func conversionCommand(srv *database.Server, java, jar, stage, format string, limit int64) (*exec.Cmd, error) {
	if err := conversionPlatform(srv); err != nil {
		return nil, err
	}
	if _, ok := chunker.Lookup(format); !ok {
		return nil, errors.New("unsupported format")
	}
	bwrap, _ := exec.LookPath("bwrap")
	systemd, _ := exec.LookPath("systemd-run")
	java, err := filepath.EvalSymlinks(java)
	if err != nil {
		return nil, err
	}
	b := []string{"--unshare-all", "--die-with-parent", "--new-session", "--cap-drop", "ALL", "--clearenv", "--setenv", "PATH", "/usr/bin:/bin", "--setenv", "LANG", "C.UTF-8", "--setenv", "HOME", "/work", "--ro-bind", "/usr", "/usr", "--ro-bind", "/bin", "/bin", "--ro-bind", "/lib", "/lib", "--ro-bind-try", "/lib64", "/lib64", "--dev", "/dev", "--proc", "/proc", "--size", strconv.FormatInt(limit*2, 10), "--tmpfs", "/work", "--size", "67108864", "--tmpfs", "/tmp", "--ro-bind", filepath.Join(stage, "input"), "/source", "--ro-bind", jar, "/chunker.jar", "--chdir", "/work"}
	for _, p := range []string{"/etc/ld.so.cache", "/etc/localtime"} {
		if _, e := os.Stat(p); e == nil {
			b = append(b, "--ro-bind", p, p)
		}
	}
	configs, _ := filepath.Glob("/etc/java*")
	for _, p := range configs {
		b = append(b, "--ro-bind", p, p)
	}
	b = append(b, javaRuntimeMountArgs(java, stage)...)
	managedHome := filepath.Join(filepath.Dir(jar), "java21")
	if rel, e := filepath.Rel(managedHome, java); e == nil && filepath.IsLocal(rel) {
		b = append(b, "--ro-bind", managedHome, managedHome)
	}
	b = append(b, "/bin/sh", "-c", conversionSandboxScript, "chunker", java, "-Xms64M", fmt.Sprintf("-Xmx%dM", srv.MemoryMaxMB/2), "-XX:+ExitOnOutOfMemoryError", "-Djava.io.tmpdir=/tmp", "-jar", "/chunker.jar", "-i", "/work/input", "-f", format, "-o", "/work/output")
	args := []string{"--scope", "--unit=neopanel-" + conversionID() + ".scope", "--quiet", "--property=KillMode=control-group", "--property=TasksMax=128", fmt.Sprintf("--property=MemoryMax=%dM", srv.MemoryMaxMB), "--property=MemorySwapMax=0", fmt.Sprintf("--property=CPUQuota=%d%%", srv.CPULimitPercent), "/bin/sh", "-c", conversionQuotaCheck, "chunker-limits", strconv.FormatInt(int64(srv.MemoryMaxMB)*(1<<20), 10), strconv.Itoa(srv.CPULimitPercent), bwrap}
	args = append(args, b...)
	cmd := exec.Command(systemd, args...)
	cmd.Dir = stage
	cmd.Env = []string{"PATH=/usr/sbin:/usr/bin:/sbin:/bin", "LANG=C.UTF-8"}
	return cmd, nil
}
func conversionDiskBudget(ctx context.Context, srv *database.Server, live, stage string, limit, size int64) error {
	_, used, err := chunker.Scan(ctx, live, chunker.MaxBytes*8)
	if err != nil {
		return err
	}
	retained, e := conversionRetainedBytes(ctx, filepath.Dir(live), srv.ID)
	if e != nil {
		return e
	}
	used += retained
	// Reserve real input clone + maximum published output + backup/retained source.
	reserve := limit + size*2 + (64 << 20)
	if srv.DiskLimitMB > 0 && (used > srv.DiskLimitMB*(1<<20) || reserve > srv.DiskLimitMB*(1<<20)-used) {
		return errors.New("insufficient assigned disk budget for conversion, backup and retained original")
	}
	var stat unix.Statfs_t
	if err = unix.Statfs(stage, &stat); err != nil {
		return err
	}
	if uint64(stat.Bavail)*uint64(stat.Bsize) < uint64(reserve) {
		return errors.New("insufficient free disk for conversion")
	}
	return nil
}

// Atomic exchange never exposes a missing live world, including a daemon crash.
// Failure leaves both trees untouched. The prior world remains under safetyName.
func conversionExchange(live, world, fresh, safetyName string) error {
	if strings.ContainsAny(world+safetyName, "/\\") {
		return errors.New("invalid exchange path")
	}
	parent, err := os.Open(live)
	if err != nil {
		return err
	}
	defer parent.Close()
	stage, err := os.Open(filepath.Dir(fresh))
	if err != nil {
		return err
	}
	defer stage.Close()
	if err = unix.Renameat2(int(stage.Fd()), filepath.Base(fresh), int(parent.Fd()), world, unix.RENAME_EXCHANGE); err != nil {
		return fmt.Errorf("atomic world replacement unavailable; original unchanged: %w", err)
	}
	// The private fresh path now holds the original. Its retention never depends
	// on a second rename succeeding. Caller records that exact safety path.
	return nil
}
