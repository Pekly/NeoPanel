//go:build !windows

package rooted

import "syscall"

const nonblock = syscall.O_NONBLOCK
