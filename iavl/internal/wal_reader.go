package internal

import (
	"fmt"
	"os"
)

func NewWALReader(file *os.File) (*WALReader, error) {
	kvr, err := NewKVDataReader(file)
	if err != nil {
		return nil, err
	}
	if kvr.Len() == 0 || kvr.At(0) != byte(KVEntryWALStart) {
		return nil, fmt.Errorf("data does not contain a valid WAL start entry")
	}
	startVersion, bytesRead, err := kvr.readVarint(1)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAL start layer: %w", err)
	}
	return &WALReader{
		rdr:         kvr,
		offset:      bytesRead,
		Version:     startVersion,
		keyMappings: make(map[int][]byte),
	}, nil
}

// WALReader reads WAL entries from a KVDataReader.
// Call Next() to read the next entry and read the Key, Value and Version fields directly as needed.
type WALReader struct {
	rdr         *KVDataReader
	offset      int
	keyMappings map[int][]byte

	// Version is the version of the WAL entries currently being read.
	Version uint64
	// Key is the key of the current WAL entry. This is valid for Set and Delete entries.
	Key []byte

	// Value is the value of the current WAL entry.
	// This is only valid for Set entries.
	Value []byte
}

// Next reads the next WAL entry, skipping over any blob entries.
// It returns the entry type, a boolean indicating if an entry was read and an error if any.
// If no more entries are available, ok will be false.
// It should only be expected that Set, Delete and Commit entries are returned.
func (wr *WALReader) Next() (entryType KVEntryType, ok bool, err error) {
	for {
		entryType, ok, err = wr.next()
		if !ok || err != nil {
			return entryType, ok, err
		}

		// skip over all blob entries, otherwise return
		switch entryType {
		case KVEntryKeyBlob, KVEntryValueBlob:
			continue
		default:
			return entryType, ok, err
		}
	}
}

// RewriteWAL rewrites the WAL entries to the given WALWriter, truncating any entries before the given version.
func RewriteWAL(writer *WALWriter, walFile *os.File, truncateBeforeVersion uint64) error {
	wr, err := NewWALReader(walFile)
	if err != nil {
		return err
	}

	startVersion := wr.Version
	newStartVersion := startVersion
	if startVersion < truncateBeforeVersion {
		newStartVersion = truncateBeforeVersion
	}
	err = writer.StartVersion(newStartVersion)
	if err != nil {
		return err
	}

	for {
		entryType, ok, err := wr.next()
		if err != nil {
			return err
		}
		if !ok {
			return nil
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
				return err
			}
		case KVEntryWALDelete:
			err = writer.WriteWALDelete(wr.Key)
			if err != nil {
				return err
			}
		case KVEntryWALSet:
			_, _, err = writer.WriteWALSet(wr.Key, wr.Value)
			if err != nil {
				return err
			}
		}
	}

}

func (wr *WALReader) next() (entryType KVEntryType, ok bool, err error) {
	// check for end of data
	if wr.offset >= wr.rdr.Len() {
		return 0, false, nil
	}

	entryType = KVEntryType(wr.rdr.At(wr.offset))
	wr.offset++
	switch entryType {
	case KVEntryWALSet:
		err := wr.readKey()
		if err != nil {
			return 0, false, err
		}

		err = wr.readValue()
		if err != nil {
			return 0, false, err
		}
	case KVEntryWALSet | KVFlagCachedKey:
		err := wr.readCachedKey()
		if err != nil {
			return 0, false, err
		}

		err = wr.readValue()
		if err != nil {
			return 0, false, err
		}
	case KVEntryWALDelete:
		err := wr.readKey()
		if err != nil {
			return 0, false, err
		}
	case KVEntryWALDelete | KVFlagCachedKey:
		err := wr.readCachedKey()
		if err != nil {
			return 0, false, err
		}
	case KVEntryWALCommit:
		var bytesRead int
		wr.Version, bytesRead, err = wr.rdr.readVarint(wr.offset)
		if err != nil {
			return 0, false, fmt.Errorf("failed to read WAL commit layer at offset %d: %w", wr.offset, err)
		}
		wr.offset += bytesRead
	case KVEntryKeyBlob:
		err = wr.readKey()
		if err != nil {
			return 0, false, fmt.Errorf("failed to read key blob at offset %d: %w", wr.offset, err)
		}
	case KVEntryValueBlob:
		var bytesRead int
		_, bytesRead, err = wr.rdr.unsafeReadBlob(wr.offset)
		if err != nil {
			return 0, false, fmt.Errorf("failed to read blob at offset %d: %w", wr.offset, err)
		}

		wr.offset += bytesRead
	default:
		return 0, false, fmt.Errorf("invalid KV entry type %d at offset %d", entryType, wr.offset-1)
	}
	return entryType, true, nil
}

func (wr *WALReader) readKey() error {
	var bytesRead int
	var err error
	wr.Key, bytesRead, err = wr.rdr.unsafeReadBlob(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read WAL key at offset %d: %w", wr.offset, err)
	}
	// cache the key
	wr.keyMappings[wr.offset] = wr.Key
	wr.offset += bytesRead
	return nil
}

func (wr *WALReader) readCachedKey() error {
	cachedKeyOffset, err := wr.rdr.readLEU32(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read cached key offset at %d: %w", wr.offset, err)
	}
	wr.offset += 4
	var ok bool
	wr.Key, ok = wr.keyMappings[int(cachedKeyOffset)]
	if !ok {
		return fmt.Errorf("cached key not found at offset %d", cachedKeyOffset)
	}
	return nil
}

func (wr *WALReader) readValue() error {
	var n int
	var err error
	wr.Value, n, err = wr.rdr.unsafeReadBlob(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read WAL value at offset %d: %w", wr.offset, err)
	}
	wr.offset += n
	return nil
}
