//go:build !linux

package runner

import (
	"context"
	"errors"
	"github.com/Pekly/NeoPanel/internal/database"
	"os/exec"
)

var errConversionPlatform = errors.New("world conversion is unavailable on this host: assigned CPU and filesystem isolation require Linux systemd cgroup v2 + Bubblewrap; Windows Job CPU enforcement is not yet implemented (no unrestricted fallback)")

func conversionPlatform(*database.Server) error { return errConversionPlatform }
func conversionCommand(*database.Server, string, string, string, string, int64) (*exec.Cmd, error) {
	return nil, errConversionPlatform
}
func conversionDiskBudget(context.Context, *database.Server, string, string, int64, int64) error {
	return errConversionPlatform
}
func conversionExchange(string, string, string, string) error { return errConversionPlatform }
