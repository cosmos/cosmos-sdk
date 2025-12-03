package internal

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"unsafe"
)

type KVDataWriter struct {
	*FileWriter
	keyCache map[string]uint32
}

func NewKVDataWriter(file *os.File) *KVDataWriter {
	fw := NewFileWriter(file)
	return &KVDataWriter{
		FileWriter: fw,
		keyCache:   make(map[string]uint32),
	}
}

func (kvs *KVDataWriter) StartWAL(version uint64) error {
	err := kvs.writeType(KVEntryWALStart)
	if err != nil {
		return err
	}
	return kvs.writeVarUint(version)
}

func (kvs *KVDataWriter) WriteUpdates(updates []KVUpdate) error {
	for _, update := range updates {
		if deleteKey := update.DeleteKey; deleteKey != nil {
			cachedOffset, cached := kvs.keyCache[unsafeBytesToString(deleteKey)]
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
				_, err := kvs.writeLenPrefixedBytes(deleteKey)
				if err != nil {
					return err
				}
			}
		} else if memNode := update.SetNode; memNode != nil {
			key := memNode.key
			cachedOffset, cached := kvs.keyCache[unsafeBytesToString(key)]
			typ := KVEntryWALSet
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
				memNode.keyOffset = keyOffset
				kvs.keyCache[unsafeBytesToString(key)] = keyOffset
			}

			valueOffset, err := kvs.writeLenPrefixedBytes(memNode.value)
			if err != nil {
				return err
			}
			memNode.valueOffset = valueOffset
		} else {
			return fmt.Errorf("invalid update: neither SetNode nor DeleteKey is set")
		}
	}
	return nil
}

func (kvs *KVDataWriter) WriteCommit(version uint64) error {
	err := kvs.writeType(KVEntryWALCommit)
	if err != nil {
		return err
	}

	return kvs.writeVarUint(version)
}

func (kvs *KVDataWriter) WriteKey(key []byte) (offset uint32, err error) {
	if offset, found := kvs.keyCache[unsafeBytesToString(key)]; found {
		return offset, nil
	}

	offset, err = kvs.writeBlob(key)
	if err != nil {
		return 0, err
	}

	kvs.keyCache[unsafeBytesToString(key)] = offset

	return offset, nil
}

func (kvs *KVDataWriter) WriteKeyValue(key, value []byte) (keyOffset, branchOffset uint32, err error) {
	keyOffset, err = kvs.WriteKey(key)
	if err != nil {
		return 0, 0, err
	}

	branchOffset, err = kvs.writeBlob(value)
	if err != nil {
		return 0, 0, err
	}

	return keyOffset, branchOffset, nil
}

func (kvs *KVDataWriter) writeBlob(bz []byte) (offset uint32, err error) {
	err = kvs.writeType(KVEntryBlob)
	if err != nil {
		return 0, err
	}
	offset, err = kvs.writeLenPrefixedBytes(bz)
	if err != nil {
		return 0, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) writeType(x KVEntryType) error {
	_, err := kvs.Write([]byte{byte(x)})
	return err
}

func (kvs *KVDataWriter) writeLenPrefixedBytes(key []byte) (offset uint32, err error) {
	lenKey := len(key)
	err = kvs.writeVarUint(uint64(lenKey))
	if err != nil {
		return 0, err
	}

	sz := kvs.Size()
	if sz > math.MaxUint32 {
		return 0, fmt.Errorf("file size overflows uint32: %d", sz)
	}
	offset = uint32(sz)

	// write key bytes
	_, err = kvs.Write(key)
	if err != nil {
		return offset, err
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
