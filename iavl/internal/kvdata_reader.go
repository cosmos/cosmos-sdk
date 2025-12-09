package internal

import (
	"encoding/binary"
	"fmt"
	"os"
)

// KVDataReader reads data from the key-value data file which can serve as a write-ahead log (WAL)
// and blob storage for keys and values.
type KVDataReader struct {
	*Mmap
}

// NewKVDataReader creates a new KVDataReader.
func NewKVDataReader(file *os.File) (*KVDataReader, error) {
	mmap, err := NewMmap(file)
	if err != nil {
		return nil, err
	}
	return &KVDataReader{
		Mmap: mmap,
	}, nil
}

// HasWAL checks if the KV data starts with a valid WAL start entry.
// It returns true and the start version if a valid WAL start entry is found.
// If not, it returns false and zero.
func (kvr *KVDataReader) HasWAL() (ok bool, startVersion uint64) {
	var err error
	ok, startVersion, _, err = kvr.hasWAL()
	if err != nil {
		return false, 0
	}
	return ok, startVersion
}

func (kvr *KVDataReader) hasWAL() (ok bool, startVersion uint64, bytesRead int, err error) {
	if kvr.Len() == 0 || kvr.At(0) != byte(KVEntryWALStart) {
		return false, 0, 0, nil
	}
	startVersion, bytesRead, err = kvr.readVarint(1)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to read WAL start version: %w", err)
	}
	return true, startVersion, bytesRead + 1, nil
}

// ReadWAL returns a WALReader to read WAL entries from the KV data.
// If the data does not start with a valid WAL start entry, an error is returned.
func (kvr *KVDataReader) ReadWAL() (*WALReader, error) {
	haveWal, startVersion, bytesRead, err := kvr.hasWAL()
	if err != nil {
		return nil, err
	}
	if !haveWal {
		return nil, fmt.Errorf("data does not contain a valid WAL start entry")
	}
	return &WALReader{
		rdr:         kvr,
		offset:      bytesRead,
		Version:     startVersion,
		keyMappings: make(map[int][]byte),
	}, nil
}

// UnsafeReadBlob reads a blob from the KV data at the given offset.
// It is expected that the blob is prefixed with its size as a varint.
// This function can be used to read any sort of key or value blob whether or not it is part of a WAL entry.
// However, this function doesn't do any checking to ensure that the offset does indeed point to a valid blob.
// The returned byte slice is unsafe and should not be used after the underlying mmap is closed.
// If it is to be retained longer, it should be copied first.
func (kvr *KVDataReader) UnsafeReadBlob(offset int) ([]byte, error) {
	bz, _, err := kvr.unsafeReadBlob(offset)
	return bz, err
}

func (kvr *KVDataReader) unsafeReadBlob(offset int) ([]byte, int, error) {
	// Read size prefix
	size, bytesRead, err := kvr.readVarint(offset)
	if err != nil {
		return nil, 0, err
	}

	// Read blob data
	offset += bytesRead
	bz, err := kvr.UnsafeSlice(offset, int(size))
	if err != nil {
		return nil, 0, err
	}

	return bz, int(size) + bytesRead, nil
}

func (kvr *KVDataReader) readVarint(offset int) (varint uint64, bytesRead int, err error) {
	_, bz, err := kvr.UnsafeSliceVar(offset, binary.MaxVarintLen64)
	if err != nil {
		return 0, 0, err
	}
	varint, bytesRead = binary.Uvarint(bz)
	if bytesRead <= 0 {
		return 0, 0, fmt.Errorf("failed to read varint at offset %d", offset)
	}
	return varint, bytesRead, nil
}

func (kvr *KVDataReader) readLEU32(offset int) (uint32, error) {
	bz, err := kvr.UnsafeSlice(offset, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bz), nil
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
			return 0, false, fmt.Errorf("failed to read WAL commit version at offset %d: %w", wr.offset, err)
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
