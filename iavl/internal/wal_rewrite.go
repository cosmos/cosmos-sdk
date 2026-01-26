package internal

import (
	"fmt"
	"os"
)

type WALRewriteInfo struct {
	KeyOffsetRemapping   map[uint64]uint64
	ValueOffsetRemapping map[uint64]uint64
}

// RewriteWAL rewrites the WAL entries to the given WALWriter, truncating any entries before the given version.
// Returns a mapping from old value offsets to new value offsets (both raw uint64 without location flags).
func RewriteWAL(writer *WALWriter, walFile *os.File, truncateBeforeVersion uint64) (*WALRewriteInfo, error) {
	var startVersion uint64
	info := &WALRewriteInfo{
		KeyOffsetRemapping:   make(map[uint64]uint64),
		ValueOffsetRemapping: make(map[uint64]uint64),
	}

	for entry, err := range ReadWAL(walFile) {
		if err != nil {
			return nil, err
		}
		if entry.Version < truncateBeforeVersion {
			continue
		}
		if startVersion == 0 {
			startVersion = entry.Version
			err = writer.StartVersion(startVersion)
			if err != nil {
				return nil, err
			}
		}

		switch entry.Op {
		case WALOpCommit:
			err = writer.WriteWALCommit(entry.Version, entry.Checkpoint)
			if err != nil {
				return nil, err
			}
		case WALOpSet:
			oldKeyOffset := uint64(entry.KeyOffset)
			oldValueOffset := uint64(entry.ValueOffset)
			newKeyOffset, newValueOffset, err := writer.WriteWALSet(entry.Key.UnsafeBytes(), entry.Value.UnsafeBytes())
			if err != nil {
				return nil, err
			}
			info.KeyOffsetRemapping[oldKeyOffset] = newKeyOffset
			info.ValueOffsetRemapping[oldValueOffset] = newValueOffset
		case WALOpDelete:
			err := writer.WriteWALDelete(entry.Key.UnsafeBytes())
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected WAL op type: %d", entry.Op)
		}
	}
	return info, nil
}
