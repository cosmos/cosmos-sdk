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

func (kvr *KVDataReader) readLEU40(offset int) (uint64, error) {
	bz, err := kvr.UnsafeSlice(offset, 5)
	if err != nil {
		return 0, err
	}
	return uint64(bz[0]) | uint64(bz[1])<<8 | uint64(bz[2])<<16 | uint64(bz[3])<<24 | uint64(bz[4])<<32, nil
}
