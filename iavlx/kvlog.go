package iavlx

import (
	"encoding/binary"
	"os"
)

const (
	KVLogEntryTypeSet byte = iota
	KVLogEntryTypeDelete
	KVLogEntryTypeCommit
	KVLogEntryTypeExtraK
	KVLogEntryTypeExtraKV
)

type KVLog struct {
	*MmapFile
}

func NewKVLog(file *os.File) (*KVLog, error) {
	mmap, err := NewMmapFile(file)
	if err != nil {
		return nil, err
	}
	return &KVLog{
		MmapFile: mmap,
	}, nil
}

func (kvs *KVLog) UnsafeReadK(offset uint32) (key []byte, err error) {
	bz, err := kvs.UnsafeSliceExact(int(offset), 4)
	if err != nil {
		return nil, err
	}
	lenKey := binary.LittleEndian.Uint32(bz)

	return kvs.UnsafeSliceExact(int(offset)+4, int(lenKey))
}

func (kvs *KVLog) UnsafeReadKV(offset uint32) (key, value []byte, err error) {
	key, err = kvs.UnsafeReadK(offset)
	if err != nil {
		return nil, nil, err
	}

	value, err = kvs.UnsafeReadK(offset + 4 + uint32(len(key)))
	if err != nil {
		return nil, nil, err
	}
	return key, value, nil
}
