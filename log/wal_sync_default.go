//go:build !linux

package log

import "os"

// walSync falls back to full fsync on non-Linux platforms.
func walSync(f *os.File) error {
	return f.Sync()
}
