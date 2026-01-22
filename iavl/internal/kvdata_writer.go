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
	keyCache sync.Map
	isKVData bool // true if writing to KV data file, false if writing to WAL
}

// NewKVDataWriter creates a new KVDataWriter.
// If isKVData is true, offsets will point to KV data file; otherwise they point to WAL.
func NewKVDataWriter(file *os.File, isKVData bool) *KVDataWriter {
	fw := NewFileWriter(file)
	return &KVDataWriter{
		FileWriter: fw,
		isKVData:   isKVData,
	}
}

const (
	// MaxKeySize defines the maximum size of a key in bytes.
	MaxKeySize = 1<<16 - 1 // 65535 bytes
	// MaxValueSize defines the maximum size of a value in bytes.
	MaxValueSize = 1<<24 - 1 // 16777215 bytes
)

// WriteKeyBlob writes a key blob and returns its offset in the file.
// This should be used for writing keys outside of WAL entries to take advantage of key caching.
func (kvs *KVDataWriter) WriteKeyBlob(key []byte) (offset KVOffset, err error) {
	if len(key) > MaxKeySize {
		return KVOffset{}, fmt.Errorf("key size exceeds maximum of %d bytes: %d bytes", MaxKeySize, len(key))
	}

	if offsetAny, found := kvs.keyCache.Load(unsafeBytesToString(key)); found {
		return offsetAny.(KVOffset), nil
	}

	offset, err = kvs.writeBlob(KVEntryKeyBlob, key)
	if err != nil {
		return KVOffset{}, err
	}

	kvs.addKeyToCache(key, offset)

	return offset, nil
}

// WriteKeyValueBlobs writes a key blob and a value blob and returns their offsets in the file.
// This should be used for writing key-value pairs in changesets where the WAL has been dropped.
func (kvs *KVDataWriter) WriteKeyValueBlobs(key, value []byte) (keyOffset, valueOffset KVOffset, err error) {
	keyOffset, err = kvs.WriteKeyBlob(key)
	if err != nil {
		return KVOffset{}, KVOffset{}, err
	}

	if len(value) > MaxValueSize {
		return KVOffset{}, KVOffset{}, fmt.Errorf("value size exceeds maximum of %d bytes: %d bytes", MaxValueSize, len(value))
	}

	valueOffset, err = kvs.writeBlob(KVEntryValueBlob, value)
	if err != nil {
		return KVOffset{}, KVOffset{}, err
	}

	return keyOffset, valueOffset, nil
}

func (kvs *KVDataWriter) writeBlob(blobType KVEntryType, bz []byte) (offset KVOffset, err error) {
	err = kvs.writeType(blobType)
	if err != nil {
		return KVOffset{}, err
	}
	offset, err = kvs.writeLenPrefixedBytes(bz)
	if err != nil {
		return KVOffset{}, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) addKeyToCache(key []byte, offset KVOffset) {
	const minCacheKeyLen = 5 // we choose 4 because offsets are uint40 (5 bytes)
	if len(key) < minCacheKeyLen {
		// don't cache very small keys
		return
	}
	kvs.keyCache.Store(unsafeBytesToString(key), offset)
}

func (kvs *KVDataWriter) writeType(x KVEntryType) error {
	_, err := kvs.Write([]byte{byte(x)})
	return err
}

func (kvs *KVDataWriter) writeLenPrefixedBytes(bz []byte) (offset KVOffset, err error) {
	// TODO: should we limit the max size of bz?
	// for keys we should probably never have anything bigger than 2^16 bytes,
	// and for values maybe 2^24 bytes?
	sz := kvs.Size()
	if sz > MaxKVOffset {
		return KVOffset{}, fmt.Errorf("file size overflows KVOffset max (%d): %d", MaxKVOffset, sz)
	}
	offset = NewKVOffset(uint64(sz), kvs.isKVData)

	lenKey := len(bz)
	err = kvs.writeVarUint(uint64(lenKey))
	if err != nil {
		return KVOffset{}, err
	}

	// write bytes
	_, err = kvs.Write(bz)
	if err != nil {
		return KVOffset{}, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) writeVarUint(x uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], x)
	_, err := kvs.Write(buf[0:n])
	return err
}

func (kvs *KVDataWriter) writeLEU40(x KVOffset) error {
	_, err := kvs.Write(x[:])
	return err
}

// unsafeBytesToString converts a byte slice to a string without allocation.
// This should be used with caution and only when the byte slice is not modified.
// But generally when we are storing a byte slice as a key in a map, this is what we should use.
func unsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
