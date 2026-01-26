package internal

import (
	"fmt"
	"os"
)

// RewriteWAL rewrites the WAL entries to the given WALWriter, truncating any entries before the given version.
// Returns a mapping from old value offsets to new value offsets (both raw uint64 without location flags).
func RewriteWAL(writer *WALWriter, walFile *os.File, truncateBeforeVersion uint64) (valueOffsetRemapping map[uint64]uint64, err error) {
	wr, err := NewWALReader(walFile)
	if err != nil {
		return nil, err
	}

	startVersion := wr.Version
	newStartVersion := startVersion
	if startVersion < truncateBeforeVersion {
		newStartVersion = truncateBeforeVersion
	}
	err = writer.StartVersion(newStartVersion)
	if err != nil {
		return nil, err
	}

	valueOffsetRemapping = make(map[uint64]uint64)
	for {
		entryType, ok, err := wr.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			return valueOffsetRemapping, nil
		}

		if wr.Version < truncateBeforeVersion {
			continue
		}

		switch entryType {
		case WALEntryStart:
			// skip start entry
			continue
		case WALEntryCommit:
			err = writer.WriteWALCommit(wr.Version)
			if err != nil {
				return nil, err
			}
		case WALEntryDelete:
			err = writer.WriteWALDelete(wr.Key.UnsafeBytes())
			if err != nil {
				return nil, err
			}
		case WALEntrySet:
			oldValueOffset := uint64(wr.setValueOffset)
			_, newValueOffset, err := writer.WriteWALSet(wr.Key.UnsafeBytes(), wr.Value.UnsafeBytes())
			if err != nil {
				return nil, err
			}
			valueOffsetRemapping[oldValueOffset] = newValueOffset
		default:
			return nil, fmt.Errorf("unexpected WAL entry type: %d", entryType)
		}
	}
}
