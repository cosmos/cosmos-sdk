package internal

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

type KVDataWriter struct {
	*FileWriter
}

func NewKVDataWriter(file *os.File) *KVDataWriter {
	fw := NewFileWriter(file)
	return &KVDataWriter{
		FileWriter: fw,
	}
}

func (kvs *KVDataWriter) WriteK(key []byte) (offset uint64, err error) {
	_, err = kvs.Write([]byte{KVDataEntryTypeExtraK})
	if err != nil {
		return offset, err
	}

	return kvs.writeLenPrefixedBytes(key)
}

func (kvs *KVDataWriter) WriteKV(key, value []byte) (offset uint32, err error) {
	_, err = kvs.Write([]byte{KVDataEntryTypeExtraKV})
	if err != nil {
		return offset, err
	}

	offset, err = kvs.writeLenPrefixedBytes(key)
	if err != nil {
		return 0, err
	}
	_, err = kvs.writeLenPrefixedBytes(value)
	return offset, err
}

func (kvs *KVDataWriter) WriteUpdates(updates []KVUpdate) error {
	for _, update := range updates {
		if deleteKey := update.DeleteKey; deleteKey != nil {
			_, err := kvs.Write([]byte{KVDataEntryTypeDelete})
			if err != nil {
				return err
			}
			_, err = kvs.writeLenPrefixedBytes(deleteKey)
			if err != nil {
				return err
			}
		} else if memNode := update.SetNode; memNode != nil {
			_, err := kvs.Write([]byte{KVDataEntryTypeSet})
			if err != nil {
				return err
			}
			offset, err := kvs.writeLenPrefixedBytes(memNode.key)
			if err != nil {
				return err
			}
			memNode.kvOffset = offset

			_, err = kvs.writeLenPrefixedBytes(memNode.value)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid update: neither SetNode nor DeleteKey is set")
		}
	}
	return nil
}

func (kvs *KVDataWriter) WriteCommit(version uint32) error {
	_, err := kvs.Write([]byte{KVDataEntryTypeCommit})
	if err != nil {
		return err
	}

	return kvs.writeLEU32(version)
}

func (kvs *KVDataWriter) writeLenPrefixedBytes(key []byte) (offset uint64, err error) {
	lenKey := len(key)
	if lenKey > math.MaxUint32 {
		return 0, fmt.Errorf("key too large: %d bytes", lenKey)
	}

	offset = uint64(kvs.Size())

	// write little endian uint32 length prefix
	err = kvs.writeLEU32(uint32(lenKey))
	if err != nil {
		return offset, err
	}

	// write key bytes
	_, err = kvs.Write(key)
	if err != nil {
		return offset, err
	}

	return offset, nil
}

func (kvs *KVDataWriter) writeLEU32(x uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], x)
	_, err := kvs.Write(buf[:])
	return err
}
