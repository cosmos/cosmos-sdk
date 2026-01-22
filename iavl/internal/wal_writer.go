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
	if cached {
		keyOffset = keyOffsetAny.(KVOffset)
	}
	// Only use cached key flag if it saves space (offset is 5 bytes)
	useCachedFlag := cached && len(key) >= 5
	typ := KVEntryWALSet
	if useCachedFlag {
		typ |= KVFlagCachedKey
	}
	err = kvs.writer.writeType(typ)
	if err != nil {
		return KVOffset{}, KVOffset{}, err
	}

	if useCachedFlag {
		err = kvs.writer.writeLEU40(keyOffset)
		if err != nil {
			return KVOffset{}, KVOffset{}, err
		}
	} else {
		// Write key inline; for short cached keys this duplicates data but saves WAL space
		newOffset, err := kvs.writer.writeLenPrefixedBytes(key)
		if err != nil {
			return KVOffset{}, KVOffset{}, err
		}
		if !cached {
			keyOffset = newOffset
			kvs.writer.addKeyToCache(key, keyOffset)
		}
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
	var cachedOffset KVOffset
	if cached {
		cachedOffset = cachedOffsetAny.(KVOffset)
	}
	// Only use cached key flag if it saves space (offset is 5 bytes)
	useCachedFlag := cached && len(key) >= 5
	typ := KVEntryWALDelete
	if useCachedFlag {
		typ |= KVFlagCachedKey
	}
	err := kvs.writer.writeType(typ)
	if err != nil {
		return err
	}

	if useCachedFlag {
		err = kvs.writer.writeLEU40(cachedOffset)
		if err != nil {
			return err
		}
	} else {
		// Write key inline; for short cached keys this duplicates data but saves WAL space
		keyOffset, err := kvs.writer.writeLenPrefixedBytes(key)
		if err != nil {
			return err
		}
		if !cached {
			kvs.writer.addKeyToCache(key, keyOffset)
		}
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
