package internal

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"unsafe"
)

// KVDataWriter writes data to a key-value data file which can be either
// a WAL file or a KV data blob file.
type KVDataWriter struct {
	*FileWriter
	keyCache sync.Map // map[string]uint64 - raw offsets without location flag
}

// NewKVDataWriter creates a new KVDataWriter.
// If isKVData is true, offsets will point to KV data file; otherwise they point to WAL.
func NewKVDataWriter(file *os.File) *KVDataWriter {
	fw := NewFileWriter(file)
	return &KVDataWriter{
		FileWriter: fw,
	}
}

const (
	// MaxKeySize defines the maximum size of a key in bytes.
	MaxKeySize = 1<<16 - 1 // 65535 bytes
	// MaxValueSize defines the maximum size of a value in bytes.
	MaxValueSize = 1<<24 - 1 // 16777215 bytes
)

// WriteKeyBlob writes a key blob and returns its raw offset in the file.
// This should be used for writing keys outside of WAL entries to take advantage of key caching.
// Use IsKVData() to determine the location flag when constructing a KVOffset.
func (kvs *KVDataWriter) WriteKeyBlob(key []byte) (offset uint64, err error) {
	if len(key) > MaxKeySize {
		return 0, fmt.Errorf("key size exceeds maximum of %d bytes: %d bytes", MaxKeySize, len(key))
	}

	if offsetAny, found := kvs.keyCache.Load(unsafeBytesToString(key)); found {
		return offsetAny.(uint64), nil
	}

	offset, err = kvs.writeLenPrefixedBytes(key)
	if err != nil {
		return 0, err
	}

	kvs.addKeyToCache(key, offset)

	return offset, nil
}

// WriteKeyValueBlobs writes a key blob and a value blob and returns their raw offsets in the file.
// This should be used for writing key-value pairs in changesets where the WAL has been dropped.
// Use IsKVData() to determine the location flag when constructing KVOffsets.
func (kvs *KVDataWriter) WriteKeyValueBlobs(key, value []byte) (keyOffset, valueOffset uint64, err error) {
	keyOffset, err = kvs.WriteKeyBlob(key)
	if err != nil {
		return 0, 0, err
	}

	if len(value) > MaxValueSize {
		return 0, 0, fmt.Errorf("value size exceeds maximum of %d bytes: %d bytes", MaxValueSize, len(value))
	}

	valueOffset, err = kvs.writeLenPrefixedBytes(value)
	if err != nil {
		return 0, 0, err
	}

	return keyOffset, valueOffset, nil
}

// addKeyToCache caches the key's raw offset for location tracking.
// All keys are cached regardless of length so we can always look up their location.
// Note: When writing WAL entries, only use WALFlagCachedKey for keys >= 5 bytes
// since the offset reference itself is 5 bytes (no space savings for shorter keys).
func (kvs *KVDataWriter) addKeyToCache(key []byte, offset uint64) {
	kvs.keyCache.Store(unsafeBytesToString(key), offset)
}

func (kvs *KVDataWriter) writeLenPrefixedBytes(bz []byte) (offset uint64, err error) {
	sz := kvs.Size()
	if sz > MaxUint40 {
		return 0, fmt.Errorf("file size overflows KVOffset max (%d): %d", MaxUint40, sz)
	}
	offset = uint64(sz)

	err = kvs.writeVarUint(uint64(len(bz)))
	if err != nil {
		return 0, err
	}

	_, err = kvs.Write(bz)
	if err != nil {
		return 0, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) writeVarUint(x uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], x)
	_, err := kvs.Write(buf[0:n])
	return err
}

func (kvs *KVDataWriter) writeLEU40(x uint64) error {
	var buf [5]byte
	buf[0] = byte(x)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	buf[4] = byte(x >> 32)
	_, err := kvs.Write(buf[:])
	return err
}

func (kvs *KVDataWriter) WriteValueBlob(value []byte) (offset uint64, err error) {
	if len(value) > MaxValueSize {
		return 0, fmt.Errorf("value size exceeds maximum of %d bytes: %d bytes", MaxValueSize, len(value))
	}

	return kvs.writeLenPrefixedBytes(value)
}

// unsafeBytesToString converts a byte slice to a string without allocation.
// This should be used with caution and only when the byte slice is not modified.
// But generally when we are storing a byte slice as a key in a map, this is what we should use.
func unsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
