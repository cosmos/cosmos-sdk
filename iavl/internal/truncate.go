package internal

import (
	"fmt"
	"os"
)

func RollbackFileToOffset(f *os.File, offset int64) error {
	filename := f.Name()
	f2, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file %s for rollback: %w", filename, err)
	}
	err = f2.Truncate(int64(offset))
	if err != nil {
		return fmt.Errorf("failed to truncate file %s to offset %d: %w", filename, offset, err)
	}
	err = f2.Close()
	if err != nil {
		return fmt.Errorf("failed to close file %s after rollback: %w", filename, err)
	}
	return nil
}
