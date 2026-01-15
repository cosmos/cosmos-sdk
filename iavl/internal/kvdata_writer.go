package internal

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

// KVDataWriter writes data to the key-value data file which can serve as a write-ahead log (WAL)
// and blob storage for keys and values.
type KVDataWriter struct {
	*FileWriter
	keyCache map[string]Uint40
}

type WALWriter struct {
	writer *KVDataWriter
}

// NewKVDataWriter creates a new KVDataWriter.
func NewKVDataWriter(file *os.File) *KVDataWriter {
	fw := NewFileWriter(file)
	return &KVDataWriter{
		FileWriter: fw,
		keyCache:   make(map[string]Uint40),
	}
}

func NewWALWriter(file *os.File, startVersion uint64) (*WALWriter, error) {
	writer := &WALWriter{
		writer: NewKVDataWriter(file),
	}
	// TODO check that file is empty?
	err := writer.writeStartWAL(startVersion)
	if err != nil {
		return nil, err
	}
	return writer, nil
}

// writeStartWAL writes a WAL start entry with the given version.
func (kvs *WALWriter) writeStartWAL(version uint64) error {
	err := kvs.writer.writeType(KVEntryWALStart)
	if err != nil {
		return err
	}
	return kvs.writer.writeVarUint(version)
}

// WriteWALUpdates writes a batch of WAL updates.
// This can ONLY be called when the currentWriter is in WAL mode.
func (kvs *WALWriter) WriteWALUpdates(updates []KVUpdate) error {
	for _, update := range updates {
		deleteKey := update.DeleteKey
		setNode := update.SetNode
		if deleteKey != nil && setNode != nil {
			return fmt.Errorf("invalid update: both SetNode and DeleteKey are set")
		}

		if deleteKey == nil && setNode == nil {
			return fmt.Errorf("invalid update: neither SetNode nor DeleteKey is set")
		}

		if deleteKey != nil {
			err := kvs.WriteWALDelete(deleteKey)
			if err != nil {
				return err
			}
		} else { // setNode != nil
			keyOffset, valueOffset, err := kvs.WriteWALSet(setNode.key, setNode.value)
			if err != nil {
				return err
			}
			setNode.keyOffset = keyOffset
			setNode.valueOffset = valueOffset
		}
	}
	return nil
}

// WriteWALSet writes a WAL set entry for the given key and value and returns their offsets in the file.
func (kvs *WALWriter) WriteWALSet(key, value []byte) (keyOffset, valueOffset Uint40, err error) {
	keyOffset, cached := kvs.writer.keyCache[unsafeBytesToString(key)]
	typ := KVEntryWALSet
	if cached {
		typ |= KVFlagCachedKey
	}
	err = kvs.writer.writeType(typ)
	if err != nil {
		return Uint40{}, Uint40{}, err
	}

	if cached {
		err = kvs.writer.writeLEU40(keyOffset)
		if err != nil {
			return Uint40{}, Uint40{}, err
		}
	} else {
		keyOffset, err = kvs.writer.writeLenPrefixedBytes(key)
		if err != nil {
			return Uint40{}, Uint40{}, err
		}
		kvs.writer.addKeyToCache(key, keyOffset)
	}

	valueOffset, err = kvs.writer.writeLenPrefixedBytes(value)
	if err != nil {
		return Uint40{}, Uint40{}, err
	}

	return keyOffset, valueOffset, nil
}

// WriteWALDelete writes a WAL delete entry for the given key.
func (kvs *WALWriter) WriteWALDelete(key []byte) error {
	cachedOffset, cached := kvs.writer.keyCache[unsafeBytesToString(key)]
	typ := KVEntryWALDelete
	if cached {
		typ |= KVFlagCachedKey
	}
	err := kvs.writer.writeType(typ)
	if err != nil {
		return err
	}

	if cached {
		err = kvs.writer.writeLEU40(cachedOffset)
		if err != nil {
			return err
		}
	} else {
		keyOffset, err := kvs.writer.writeLenPrefixedBytes(key)
		if err != nil {
			return err
		}

		kvs.writer.addKeyToCache(key, keyOffset)
	}

	return nil
}

// WriteWALCommit writes a WAL commit entry for the given version.
func (kvs *WALWriter) WriteWALCommit(version uint64) error {
	err := kvs.writer.writeType(KVEntryWALCommit)
	if err != nil {
		return err
	}

	return kvs.writer.writeVarUint(version)
}

func (kvs *WALWriter) Sync() error {
	return kvs.writer.Sync()
}

func (kvs *WALWriter) Size() int {
	return kvs.writer.Size()
}

// LookupKeyOffset looks up the offset of the given key in the key cache.
func (kvs *WALWriter) LookupKeyOffset(key []byte) (Uint40, bool) {
	offset, found := kvs.writer.keyCache[unsafeBytesToString(key)]
	return offset, found
}

const (
	// MaxKeySize defines the maximum size of a key in bytes.
	MaxKeySize = 1<<16 - 1 // 65535 bytes
	// MaxValueSize defines the maximum size of a value in bytes.
	MaxValueSize = 1<<24 - 1 // 16777215 bytes
)

// WriteKeyBlob writes a key blob and returns its offset in the file.
// This should be used for writing keys outside of WAL entries to take advantage of key caching.
func (kvs *KVDataWriter) WriteKeyBlob(key []byte) (offset Uint40, err error) {
	if len(key) > MaxKeySize {
		return Uint40{}, fmt.Errorf("key size exceeds maximum of %d bytes: %d bytes", MaxKeySize, len(key))
	}

	if offset, found := kvs.keyCache[unsafeBytesToString(key)]; found {
		return offset, nil
	}

	offset, err = kvs.writeBlob(KVEntryKeyBlob, key)
	if err != nil {
		return Uint40{}, err
	}

	kvs.addKeyToCache(key, offset)

	return offset, nil
}

// WriteKeyValueBlobs writes a key blob and a value blob and returns their offsets in the file.
// This should be used for writing key-value pairs in changesets where the WAL has been dropped.
func (kvs *KVDataWriter) WriteKeyValueBlobs(key, value []byte) (keyOffset, valueOffset Uint40, err error) {
	keyOffset, err = kvs.WriteKeyBlob(key)
	if err != nil {
		return Uint40{}, Uint40{}, err
	}

	if len(value) > MaxValueSize {
		return Uint40{}, Uint40{}, fmt.Errorf("value size exceeds maximum of %d bytes: %d bytes", MaxValueSize, len(value))
	}

	valueOffset, err = kvs.writeBlob(KVEntryValueBlob, value)
	if err != nil {
		return Uint40{}, Uint40{}, err
	}

	return keyOffset, valueOffset, nil
}

func (kvs *KVDataWriter) writeBlob(blobType KVEntryType, bz []byte) (offset Uint40, err error) {
	err = kvs.writeType(blobType)
	if err != nil {
		return Uint40{}, err
	}
	offset, err = kvs.writeLenPrefixedBytes(bz)
	if err != nil {
		return Uint40{}, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) addKeyToCache(key []byte, offset Uint40) {
	const minCacheKeyLen = 4 // we choose 4 because offsets are uint32 (4 bytes)
	if len(key) < minCacheKeyLen {
		// don't cache very small keys
		return
	}
	kvs.keyCache[unsafeBytesToString(key)] = offset
}

func (kvs *KVDataWriter) writeType(x KVEntryType) error {
	_, err := kvs.Write([]byte{byte(x)})
	return err
}

func (kvs *KVDataWriter) writeLenPrefixedBytes(bz []byte) (offset Uint40, err error) {
	// TODO: should we limit the max size of bz?
	// for keys we should probably never have anything bigger than 2^16 bytes,
	// and for values maybe 2^24 bytes?
	sz := kvs.Size()
	if sz > MaxUint40 {
		return Uint40{}, fmt.Errorf("file size overflows uint32: %d", sz)
	}
	offset = NewUint40(uint64(sz))

	lenKey := len(bz)
	err = kvs.writeVarUint(uint64(lenKey))
	if err != nil {
		return Uint40{}, err
	}

	// write bytes
	_, err = kvs.Write(bz)
	if err != nil {
		return Uint40{}, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) writeVarUint(x uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], x)
	_, err := kvs.Write(buf[0:n])
	return err
}

func (kvs *KVDataWriter) writeLEU40(x Uint40) error {
	_, err := kvs.Write(x[:])
	return err
}

// unsafeBytesToString converts a byte slice to a string without allocation.
// This should be used with caution and only when the byte slice is not modified.
// But generally when we are storing a byte slice as a key in a map, this is what we should use.
func unsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
