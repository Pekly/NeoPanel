package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var ErrServerProcessAlive = errors.New("server processes remain; deletion and destructive maintenance refused")

type procIdentity struct {
	PID   int
	Group int
	State string
}

func parseProcIdentity(raw string) (procIdentity, error) {
	// comm is parenthesized and may itself contain spaces and ')'.
	open, close := strings.IndexByte(raw, '('), strings.LastIndexByte(raw, ')')
	if open < 1 || close <= open {
		return procIdentity{}, fmt.Errorf("invalid process stat")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(raw[:open]))
	if err != nil || pid <= 0 {
		return procIdentity{}, fmt.Errorf("invalid process PID")
	}
	fields := strings.Fields(raw[close+1:])
	if len(fields) < 3 {
		return procIdentity{}, fmt.Errorf("incomplete process stat")
	}
	group, err := strconv.Atoi(fields[2])
	if err != nil {
		return procIdentity{}, err
	}
	return procIdentity{PID: pid, Group: group, State: fields[0]}, nil
}

func liveProcIdentities(root string) ([]procIdentity, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect process table: %w", err)
	}
	var result []procIdentity
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil || !entry.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, entry.Name(), "stat"))
		if processDisappeared(err) {
			continue
		} // process exited during inspection
		if err != nil {
			return nil, fmt.Errorf("cannot inspect PID %s: %w", entry.Name(), err)
		}
		identity, err := parseProcIdentity(string(raw))
		if err != nil {
			return nil, fmt.Errorf("cannot inspect PID %s: %w", entry.Name(), err)
		}
		if identity.State != "Z" && identity.State != "X" {
			result = append(result, identity)
		}
	}
	return result, nil
}

func processPathWithinServer(dir, target string) bool {
	// Linux appends this suffix to cwd/exe after an older panel unlinked a live tree.
	target = strings.TrimSuffix(target, " (deleted)")
	if !filepath.IsAbs(target) {
		return false
	}
	rel, err := filepath.Rel(dir, filepath.Clean(target))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel)
}

func inspectServerProcesses(procRoot, dir string, readlink func(string) (string, error)) error {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		absolute = resolved
	}
	processes, err := liveProcIdentities(procRoot)
	if err != nil {
		return err
	}
	for _, process := range processes {
		if process.PID == os.Getpid() {
			continue
		}
		for _, link := range []string{"cwd", "exe"} {
			target, err := readlink(filepath.Join(procRoot, strconv.Itoa(process.PID), link))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return fmt.Errorf("cannot verify shutdown: inspect PID %d %s: %w", process.PID, link, err)
			}
			if processPathWithinServer(absolute, target) {
				return fmt.Errorf("%w: PID %d still references %s (%s). Stop the verified process or its systemd scope on this node and retry; no files were removed", ErrServerProcessAlive, process.PID, dir, link)
			}
		}
	}
	return nil
}

func processDisappeared(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, syscall.ESRCH)
}

func processHasUID(status string, uid int) (bool, error) {
	for _, line := range strings.Split(status, "\n") {
		if !strings.HasPrefix(line, "Uid:") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "Uid:"))
		if len(fields) != 4 {
			return false, fmt.Errorf("incomplete process UID fields")
		}
		ours := false
		for _, field := range fields {
			value, err := strconv.Atoi(field)
			if err != nil {
				return false, err
			}
			if value == uid {
				ours = true
			}
		}
		return ours, nil
	}
	return false, fmt.Errorf("process UID fields unavailable")
}
