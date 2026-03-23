package internal

import (
	"fmt"
	"os"
)

func RollbackWAL(walFile *os.File, targetVersion uint64) error {
	var rollbackOffset int
	for entry, err := range ReadWAL(walFile) {
		if err != nil {
			return fmt.Errorf("failed to read WAL: %w", err)
		}
		if entry.Version > targetVersion {
			break
		}
		if entry.Op == WALOpCommit {
			rollbackOffset = entry.EndOffset
		}
	}

	if rollbackOffset == 0 {
		return fmt.Errorf("no commit entry found in WAL at or before target version %d", targetVersion)
	}

	return RollbackFileToOffset(walFile, int64(rollbackOffset))
}
