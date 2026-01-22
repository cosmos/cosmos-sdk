package internal

import (
	"fmt"
	"os"
)

type WALWriter struct {
	writer       *KVDataWriter
	startVersion uint64
}

func NewWALWriter(file *os.File) *WALWriter {
	return &WALWriter{
		writer: NewKVDataWriter(file, false),
	}
}

// StartVersion should be called before writing any WAL entries for each version.
// This may or may not result in a WAL start entry being written, depending on whether
// this is the first version in the WAL or not.
func (kvs *WALWriter) StartVersion(version uint64) error {
	if kvs.startVersion != 0 {
		// start version already set
		return nil
	}
	kvs.startVersion = version
	return kvs.writeStartWAL(version)
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
func (kvs *WALWriter) WriteWALSet(key, value []byte) (keyOffset, valueOffset KVOffset, err error) {
	keyOffsetAny, cached := kvs.writer.keyCache.Load(unsafeBytesToString(key))
	typ := KVEntryWALSet
	if cached {
		typ |= KVFlagCachedKey
		keyOffset = keyOffsetAny.(KVOffset)
	}
	err = kvs.writer.writeType(typ)
	if err != nil {
		return KVOffset{}, KVOffset{}, err
	}

	if cached {
		err = kvs.writer.writeLEU40(keyOffset)
		if err != nil {
			return KVOffset{}, KVOffset{}, err
		}
	} else {
		keyOffset, err = kvs.writer.writeLenPrefixedBytes(key)
		if err != nil {
			return KVOffset{}, KVOffset{}, err
		}
		kvs.writer.addKeyToCache(key, keyOffset)
	}

	valueOffset, err = kvs.writer.writeLenPrefixedBytes(value)
	if err != nil {
		return KVOffset{}, KVOffset{}, err
	}

	return keyOffset, valueOffset, nil
}

// WriteWALDelete writes a WAL delete entry for the given key.
func (kvs *WALWriter) WriteWALDelete(key []byte) error {
	cachedOffsetAny, cached := kvs.writer.keyCache.Load(unsafeBytesToString(key))
	typ := KVEntryWALDelete
	var cachedOffset KVOffset
	if cached {
		typ |= KVFlagCachedKey
		cachedOffset = cachedOffsetAny.(KVOffset)
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
func (kvs *WALWriter) LookupKeyOffset(key []byte) (KVOffset, bool) {
	offset, found := kvs.writer.keyCache.Load(unsafeBytesToString(key))
	if found {
		return offset.(KVOffset), true
	}
	return KVOffset{}, false
}
