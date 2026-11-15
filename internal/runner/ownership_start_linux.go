//go:build linux

package runner

import (
	"os"
	"os/exec"
)

func startOwnedCommand(cmd *exec.Cmd, lock *os.File, memoryMB int64) (*ownedCommand, error) {
	configureProcessOwnership(cmd, lock)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &ownedCommand{cmd: cmd}, nil
}
