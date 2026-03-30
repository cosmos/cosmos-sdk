package internal

import (
	"fmt"
	"os"
)

// RollbackWAL truncates a WAL file to remove all entries after targetVersion.
// It scans the WAL for commit entries and truncates to the byte offset immediately after
// the last commit entry at or before targetVersion.
//
// This is used by explicit version rollback (rollbackToVersion), NOT by crash recovery
// (which uses ReplayWALForStartup's auto-repair) or in-flight commit rollback
// (which uses WALWriter.Rollback, which already knows the offset without scanning).
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
