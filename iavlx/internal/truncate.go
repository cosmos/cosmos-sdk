package internal

import (
	"fmt"
	"os"
)

// RollbackFileToOffset truncates the file to the given offset and syncs it.
// It opens a separate write handle because the passed-in file may be read-only or memory-mapped.
func RollbackFileToOffset(f *os.File, offset int64) error {
	filename := f.Name()
	f2, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file %s for rollback: %w", filename, err)
	}
	err = f2.Truncate(offset)
	if err != nil {
		return fmt.Errorf("failed to truncate file %s to offset %d: %w", filename, offset, err)
	}
	err = f2.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file %s after truncation: %w", filename, err)
	}
	err = f2.Close()
	if err != nil {
		return fmt.Errorf("failed to close file %s after rollback: %w", filename, err)
	}
	return nil
}
