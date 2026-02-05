package internal

import (
	"fmt"
	"io"
	"iter"
	"os"
)

func ReadWAL(file *os.File) iter.Seq2[WALEntry, error] {
	kvr, err := NewKVDataReader(file)
	if err != nil {
		return func(yield func(WALEntry, error) bool) {
			yield(WALEntry{}, fmt.Errorf("failed to open WAL data store: %w", err))
		}
	}
	if kvr.Len() == 0 {
		// empty WAL file — no entries to replay
		return func(yield func(WALEntry, error) bool) {}
	}
	if kvr.At(0) != byte(WALEntryStart) {
		return func(yield func(WALEntry, error) bool) {
			yield(WALEntry{}, fmt.Errorf("data does not contain a valid WAL start entry"))
		}
	}
	startVersion, bytesRead, err := kvr.readVarint(1)
	if err != nil {
		return func(yield func(WALEntry, error) bool) {
			yield(WALEntry{}, fmt.Errorf("failed to read WAL start layer: %w", err))
		}
	}
	rdr := &walReader{
		rdr:         kvr,
		offset:      1 + bytesRead,
		version:     startVersion,
		keyMappings: make(map[int]UnsafeBytes),
	}

	return func(yield func(WALEntry, error) bool) {
		for {
			entry, err := rdr.next()
			if err != nil {
				if err == io.EOF {
					return
				} else {
					yield(WALEntry{}, err)
					return
				}
			}
			if !yield(entry, nil) {
				return
			}
		}
	}
}

// walReader reads WAL entries from a KVDataReader.
// Call Next() to read the next entry and read the Key, Value and Version fields directly as needed.
type walReader struct {
	rdr           *KVDataReader
	offset        int
	version       uint64
	keyMappings   map[int]UnsafeBytes
	key, value    UnsafeBytes
	keyOffset     int
	valueOffset   int
	lastEntryType WALEntryType
}

func (wr *walReader) next() (WALEntry, error) {
	// check for end of data
	if wr.offset >= wr.rdr.Len() {
		// WAL must end with a commit entry
		switch wr.lastEntryType {
		case WALEntryCommit, WALEntryCommit | WALFlagCheckpoint:
			return WALEntry{}, io.EOF
		default:
			return WALEntry{}, fmt.Errorf("WAL ended unexpectedly at offset %d without a commit entry, last entry type %d", wr.offset, wr.lastEntryType)
		}
	}

	entryType := WALEntryType(wr.rdr.At(wr.offset))
	wr.lastEntryType = entryType
	wr.offset++
	switch entryType {
	case WALEntrySet:
		err := wr.readKey()
		if err != nil {
			return WALEntry{}, err
		}

		err = wr.readValue()
		if err != nil {
			return WALEntry{}, err
		}

		return WALEntry{
			Op:          WALOpSet,
			Version:     wr.version,
			Key:         wr.key,
			Value:       wr.value,
			KeyOffset:   wr.keyOffset,
			ValueOffset: wr.valueOffset,
		}, nil
	case WALEntrySet | WALFlagCachedKey:
		err := wr.readCachedKey()
		if err != nil {
			return WALEntry{}, err
		}

		err = wr.readValue()
		if err != nil {
			return WALEntry{}, err
		}
		return WALEntry{
			Op:          WALOpSet,
			Version:     wr.version,
			Key:         wr.key,
			Value:       wr.value,
			KeyOffset:   wr.keyOffset,
			ValueOffset: wr.valueOffset,
		}, nil
	case WALEntryDelete:
		err := wr.readKey()
		if err != nil {
			return WALEntry{}, err
		}

		return WALEntry{
			Op:        WALOpDelete,
			Version:   wr.version,
			Key:       wr.key,
			KeyOffset: wr.keyOffset,
		}, nil
	case WALEntryDelete | WALFlagCachedKey:
		err := wr.readCachedKey()
		if err != nil {
			return WALEntry{}, err
		}

		return WALEntry{
			Op:        WALOpDelete,
			Version:   wr.version,
			Key:       wr.key,
			KeyOffset: wr.keyOffset,
		}, nil
	case WALEntryCommit, WALEntryCommit | WALFlagCheckpoint:
		checkpoint := false
		if entryType&WALFlagCheckpoint != 0 {
			checkpoint = true
		}
		var bytesRead int
		var err error
		version, bytesRead, err := wr.rdr.readVarint(wr.offset)
		if err != nil {
			err = fmt.Errorf("failed to read WAL commit layer at offset %d: %w", wr.offset, err)
			return WALEntry{}, err
		}
		wr.offset += bytesRead
		if version != wr.version {
			return WALEntry{}, fmt.Errorf("WAL commit version %d does not match current version %d at offset %d", version, wr.version, wr.offset-bytesRead)
		}
		wr.version++ // increment version for next entries

		return WALEntry{
			Op:         WALOpCommit,
			Version:    version, // version being committed
			Checkpoint: checkpoint,
		}, nil
	default:
		return WALEntry{}, fmt.Errorf("invalid KV entry type %d at offset %d", entryType, wr.offset-1)
	}
}

func (wr *walReader) readKey() error {
	wr.keyOffset = wr.offset
	bz, bytesRead, err := wr.rdr.unsafeReadBlob(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read WAL key at offset %d: %w", wr.offset, err)
	}
	wr.key = WrapUnsafeBytes(bz)
	// cache the key by its offset for cached key lookups
	wr.keyMappings[wr.keyOffset] = wr.key
	wr.offset += bytesRead
	return nil
}

func (wr *walReader) readCachedKey() error {
	cachedKeyOffset, err := wr.rdr.readLEU40(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read cached key offset at %d: %w", wr.offset, err)
	}
	wr.keyOffset = int(cachedKeyOffset)
	wr.offset += 5
	var ok bool
	wr.key, ok = wr.keyMappings[int(cachedKeyOffset)]
	if !ok {
		return fmt.Errorf("cached key not found at offset %d", cachedKeyOffset)
	}
	return nil
}

func (wr *walReader) readValue() error {
	bz, n, err := wr.rdr.unsafeReadBlob(wr.offset)
	if err != nil {
		return fmt.Errorf("failed to read WAL value at offset %d: %w", wr.offset, err)
	}
	wr.valueOffset = wr.offset
	wr.value = WrapUnsafeBytes(bz)
	wr.offset += n
	return nil
}
