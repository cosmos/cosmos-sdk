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
		offset:      1 + bytesRead,
		Version:     startVersion,
		keyMappings: make(map[int]UnsafeBytes),
	}, nil
}

// WALReader reads WAL entries from a KVDataReader.
// Call Next() to read the next entry and read the Key, Value and Version fields directly as needed.
type WALReader struct {
	rdr            *KVDataReader
	offset         int
	setValueOffset int
	keyMappings    map[int]UnsafeBytes

	// Version is the version of the WAL entries currently being read.
	Version uint64
	// Key is the key of the current WAL entry. This is valid for Set and Delete entries.
	// This is an UnsafeBytes pointing into the mmap'd WAL data - copy if needed beyond the reader's lifetime.
	Key UnsafeBytes

	// Value is the value of the current WAL entry.
	// This is only valid for Set entries.
	// This is an UnsafeBytes pointing into the mmap'd WAL data - copy if needed beyond the reader's lifetime.
	Value UnsafeBytes
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

		wr.setValueOffset = wr.offset // we save this for remapping when rewriting the WAL
		err = wr.readValue()
		if err != nil {
			return 0, false, err
		}
	case KVEntryWALSet | KVFlagCachedKey:
		err := wr.readCachedKey()
		if err != nil {
			return 0, false, err
		}

		wr.setValueOffset = wr.offset // we save this for remapping when rewriting the WAL
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
	keyOffset := wr.offset
	bz, bytesRead, err := wr.rdr.unsafeReadBlob(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read WAL key at offset %d: %w", wr.offset, err)
	}
	wr.Key = WrapUnsafeBytes(bz)
	// cache the key by its offset for cached key lookups
	wr.keyMappings[keyOffset] = wr.Key
	wr.offset += bytesRead
	return nil
}

func (wr *WALReader) readCachedKey() error {
	cachedKeyOffset, err := wr.rdr.readLEU40(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read cached key offset at %d: %w", wr.offset, err)
	}
	wr.offset += 5
	var ok bool
	// The stored offset may have location flag set, mask it off for lookup
	wr.Key, ok = wr.keyMappings[int(cachedKeyOffset&kvOffsetMask)]
	if !ok {
		return fmt.Errorf("cached key not found at offset %d", cachedKeyOffset)
	}
	return nil
}

func (wr *WALReader) readValue() error {
	bz, n, err := wr.rdr.unsafeReadBlob(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read WAL value at offset %d: %w", wr.offset, err)
	}
	wr.Value = WrapUnsafeBytes(bz)
	wr.offset += n
	return nil
}
