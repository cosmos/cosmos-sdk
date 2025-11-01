//go:build linux

package log

import (
	"os"
	"syscall"
)

// walSync uses fdatasync on Linux to sync only data (metadata may not be flushed),
// which is generally cheaper than a full fsync.
func walSync(f *os.File) error {
	return syscall.Fdatasync(int(f.Fd()))
}
