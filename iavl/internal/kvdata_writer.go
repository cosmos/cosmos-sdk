package internal

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"unsafe"
)

// KVDataWriter writes data to the key-value data file which can serve as a write-ahead log (WAL)
// and blob storage for keys and values.
type KVDataWriter struct {
	*FileWriter
	keyCache map[string]uint32
}

// NewKVDataWriter creates a new KVDataWriter.
func NewKVDataWriter(file *os.File) *KVDataWriter {
	fw := NewFileWriter(file)
	return &KVDataWriter{
		FileWriter: fw,
		keyCache:   make(map[string]uint32),
	}
}

func (kvs *KVDataWriter) WriteStartWAL(version uint64) error {
	if kvs.Size() != 0 {
		return fmt.Errorf("cannot write WAL start to non-empty file")
	}
	err := kvs.writeType(KVEntryWALStart)
	if err != nil {
		return err
	}
	return kvs.writeVarUint(version)
}

func (kvs *KVDataWriter) WriteWALUpdates(updates []KVUpdate) error {
	for _, update := range updates {
		if deleteKey := update.DeleteKey; deleteKey != nil {
			err := kvs.WriteWALDelete(deleteKey)
			if err != nil {
				return err
			}
		} else if memNode := update.SetNode; memNode != nil {
			keyOffset, valueOffset, err := kvs.WriteWALSet(memNode.key, memNode.value)
			if err != nil {
				return err
			}
			memNode.keyOffset = keyOffset
			memNode.valueOffset = valueOffset
		} else {
			return fmt.Errorf("invalid update: neither SetNode nor DeleteKey is set")
		}
	}
	return nil
}

func (kvs *KVDataWriter) WriteWALSet(key, value []byte) (keyOffset, valueOffset uint32, err error) {
	keyOffset, cached := kvs.keyCache[unsafeBytesToString(key)]
	typ := KVEntryWALSet
	if cached {
		typ |= KVFlagCachedKey
	}
	err = kvs.writeType(typ)
	if err != nil {
		return 0, 0, err
	}

	if cached {
		err = kvs.writeLEU32(keyOffset)
		if err != nil {
			return 0, 0, err
		}
	} else {
		var err error
		keyOffset, err = kvs.writeLenPrefixedBytes(key)
		if err != nil {
			return 0, 0, err
		}
		kvs.addKeyToCache(key, keyOffset)
	}

	valueOffset, err = kvs.writeLenPrefixedBytes(value)
	if err != nil {
		return 0, 0, err
	}

	return keyOffset, valueOffset, nil
}

func (kvs *KVDataWriter) WriteWALDelete(key []byte) error {
	cachedOffset, cached := kvs.keyCache[unsafeBytesToString(key)]
	typ := KVEntryWALDelete
	if cached {
		typ |= KVFlagCachedKey
	}
	err := kvs.writeType(typ)
	if err != nil {
		return err
	}

	if cached {
		err = kvs.writeLEU32(cachedOffset)
		if err != nil {
			return err
		}
	} else {
		keyOffset, err := kvs.writeLenPrefixedBytes(key)
		if err != nil {
			return err
		}

		kvs.addKeyToCache(key, keyOffset)
	}

	return nil
}

func (kvs *KVDataWriter) WriteWALCommit(version uint64) error {
	err := kvs.writeType(KVEntryWALCommit)
	if err != nil {
		return err
	}

	return kvs.writeVarUint(version)
}

func (kvs *KVDataWriter) WriteKeyBlob(key []byte) (offset uint32, err error) {
	if offset, found := kvs.keyCache[unsafeBytesToString(key)]; found {
		return offset, nil
	}

	offset, err = kvs.writeBlob(KVEntryKeyBlob, key)
	if err != nil {
		return 0, err
	}

	kvs.addKeyToCache(key, offset)

	return offset, nil
}

func (kvs *KVDataWriter) WriteKeyValueBlobs(key, value []byte) (keyOffset, valueOffset uint32, err error) {
	keyOffset, err = kvs.WriteKeyBlob(key)
	if err != nil {
		return 0, 0, err
	}

	valueOffset, err = kvs.writeBlob(KVEntryKeyBlob, value)
	if err != nil {
		return 0, 0, err
	}

	return keyOffset, valueOffset, nil
}

func (kvs *KVDataWriter) writeBlob(blobType KVEntryType, bz []byte) (offset uint32, err error) {
	err = kvs.writeType(blobType)
	if err != nil {
		return 0, err
	}
	offset, err = kvs.writeLenPrefixedBytes(bz)
	if err != nil {
		return 0, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) addKeyToCache(key []byte, offset uint32) {
	if len(key) < 4 {
		// don't cache very small keys
		return
	}
	kvs.keyCache[unsafeBytesToString(key)] = offset
}

func (kvs *KVDataWriter) writeType(x KVEntryType) error {
	_, err := kvs.Write([]byte{byte(x)})
	return err
}

func (kvs *KVDataWriter) writeLenPrefixedBytes(bz []byte) (offset uint32, err error) {
	// TODO: should we limit the max size of bz?
	// for keys we should probably never have anything bigger than 2^16 bytes,
	// and for values maybe 2^24 bytes?
	sz := kvs.Size()
	if sz > math.MaxUint32 {
		return 0, fmt.Errorf("file size overflows uint32: %d", sz)
	}
	offset = uint32(sz)

	lenKey := len(bz)
	err = kvs.writeVarUint(uint64(lenKey))
	if err != nil {
		return 0, err
	}

	// write bytes
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

func (kvs *KVDataWriter) writeLEU32(x uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], x)
	_, err := kvs.Write(buf[:])
	return err
}

func unsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
