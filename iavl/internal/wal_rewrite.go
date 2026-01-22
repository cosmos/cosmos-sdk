package internal

import (
	"fmt"
	"os"
)

// RewriteWAL rewrites the WAL entries to the given WALWriter, truncating any entries before the given version.
func RewriteWAL(writer *WALWriter, walFile *os.File, truncateBeforeVersion uint64) (valueOffsetRemapping map[KVOffset]KVOffset, err error) {
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

	valueOffsetRemapping = make(map[KVOffset]KVOffset)
	for {
		entryType, ok, err := wr.next()
		if err != nil {
			return nil, err
		}
		if !ok {
			return valueOffsetRemapping, nil
		}

		switch entryType {
		case KVEntryWALStart:
			// skip start entry
			continue
		case KVEntryKeyBlob, KVEntryValueBlob:
			// skip blob entries
			continue
		case KVEntryWALCommit:
			if wr.Version < truncateBeforeVersion {
				continue
			}
			err = writer.WriteWALCommit(wr.Version)
			if err != nil {
				return nil, err
			}
		case KVEntryWALDelete:
			err = writer.WriteWALDelete(wr.Key)
			if err != nil {
				return nil, err
			}
		case KVEntryWALSet:
			_, valueOffset, err := writer.WriteWALSet(wr.Key, wr.Value)
			valueOffsetRemapping[NewKVOffset(uint64(wr.setValueOffset), false)] = valueOffset
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected WAL entry type: %d", entryType)
		}
	}

}
