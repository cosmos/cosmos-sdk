package iavlx

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

type KVLogWriter struct {
	*FileWriter
}

func NewKVDataWriter(file *os.File) *KVLogWriter {
	fw := NewFileWriter(file)
	return &KVLogWriter{
		FileWriter: fw,
	}
}

func (kvs *KVLogWriter) WriteK(key []byte) (offset uint32, err error) {
	_, err = kvs.Write([]byte{KVLogEntryTypeExtraK})
	if err != nil {
		return offset, err
	}

	return kvs.writeLenPrefixedBytes(key)
}

func (kvs *KVLogWriter) WriteKV(key, value []byte) (offset uint32, err error) {
	_, err = kvs.Write([]byte{KVLogEntryTypeExtraKV})
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

func (kvs *KVLogWriter) WriteUpdates(updates []KVUpdate) error {
	for _, update := range updates {
		if deleteKey := update.DeleteKey; deleteKey != nil {
			_, err := kvs.Write([]byte{KVLogEntryTypeDelete})
			if err != nil {
				return err
			}
			_, err = kvs.writeLenPrefixedBytes(deleteKey)
			if err != nil {
				return err
			}
		} else if memNode := update.SetNode; memNode != nil {
			_, err := kvs.Write([]byte{KVLogEntryTypeSet})
			if err != nil {
				return err
			}
			offset, err := kvs.writeLenPrefixedBytes(memNode.key)
			if err != nil {
				return err
			}
			memNode.keyOffset = offset

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

func (kvs *KVLogWriter) WriteCommit(version uint32) error {
	_, err := kvs.Write([]byte{KVLogEntryTypeCommit})
	if err != nil {
		return err
	}

	return kvs.writeLEU32(version)
}

func (kvs *KVLogWriter) writeLenPrefixedBytes(key []byte) (offset uint32, err error) {
	lenKey := len(key)
	if lenKey > math.MaxUint32 {
		return 0, fmt.Errorf("key too large: %d bytes", lenKey)
	}

	offset = uint32(kvs.Size())

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

func (kvs *KVLogWriter) writeLEU32(x uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], x)
	_, err := kvs.Write(buf[:])
	return err
}
