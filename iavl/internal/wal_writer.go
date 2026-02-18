package internal

import (
	"context"
	"fmt"
	"os"
)

type WALWriter struct {
	writer            *KVDataWriter
	startVersion      uint64
	lastVersionOffset uint64
	// currentUpdates holds the current batch of updates that have just been written,
	// in case we need to roll back
	currentUpdates []KVUpdate
}

func NewWALWriter(file *os.File) *WALWriter {
	return &WALWriter{
		writer: NewKVDataWriter(file),
	}
}

var WALWriteAbortedErr = fmt.Errorf("WAL write aborted and rolled back")

func (kvs *WALWriter) WriteWALVersion(ctx context.Context, version uint64, updates []KVUpdate, checkpoint bool) error {
	kvs.lastVersionOffset = uint64(kvs.writer.Size())
	kvs.currentUpdates = updates

	if err := kvs.doWriteWALVersion(ctx, version, updates, checkpoint); err != nil {
		rbErr := kvs.Rollback()
		if rbErr != nil {
			return fmt.Errorf("failed to write WAL version: %w; rollback also failed: %v", err, rbErr)
		}
		return fmt.Errorf("%w; due to: %v", WALWriteAbortedErr, err)
	}
	return nil
}

func (kvs *WALWriter) doWriteWALVersion(ctx context.Context, version uint64, updates []KVUpdate, checkpoint bool) error {
	if err := kvs.writeStartVersion(version); err != nil {
		return err
	}

	if err := kvs.writeWALUpdates(ctx, updates); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return kvs.writeWALCommit(version, checkpoint)
}

// writeStartVersion writes a WAL start entry if the start version is not already set.
func (kvs *WALWriter) writeStartVersion(version uint64) error {
	if kvs.startVersion != 0 {
		// start version already set
		return nil
	}
	kvs.startVersion = version
	err := kvs.writeType(WALEntryStart)
	if err != nil {
		return err
	}
	return kvs.writer.writeVarUint(version)
}

// WriteWALUpdates writes a batch of WAL updates.
// This can ONLY be called when the currentWriter is in WAL mode.
func (kvs *WALWriter) writeWALUpdates(ctx context.Context, updates []KVUpdate) error {
	for _, update := range updates {
		if err := ctx.Err(); err != nil {
			return err
		}

		deleteKey := update.DeleteKey
		setNode := update.SetNode
		if deleteKey != nil && setNode != nil {
			return fmt.Errorf("invalid update: both SetNode and DeleteKey are set")
		}

		if deleteKey == nil && setNode == nil {
			return fmt.Errorf("invalid update: neither SetNode nor DeleteKey is set")
		}

		if deleteKey != nil {
			err := kvs.writeWALDelete(WrapSafeBytes(deleteKey))
			if err != nil {
				return err
			}
		} else { // setNode != nil
			keyOffset, valueOffset, err := kvs.writeWALSet(WrapSafeBytes(setNode.key), WrapSafeBytes(setNode.value))
			if err != nil {
				return err
			}
			setNode.walKeyOffset = keyOffset
			setNode.walValueOffset = valueOffset
		}
	}
	return nil
}

// WriteWALSet writes a WAL set entry for the given key and value and returns their raw offsets.
func (kvs *WALWriter) writeWALSet(key, value UnsafeBytes) (keyOffset, valueOffset uint64, err error) {
	unsafeKey := key.UnsafeBytes()
	// safe to use the unsafe key bytes for lookup
	keyOffsetAny, cached := kvs.writer.keyCache.Load(unsafeBytesToString(unsafeKey))
	if cached {
		keyOffset = keyOffsetAny.(uint64)
	}
	// Only use cached key flag if it saves space (offset is 5 bytes)
	useCachedFlag := cached && len(unsafeKey) >= 5
	typ := WALEntrySet
	if useCachedFlag {
		typ |= WALFlagCachedKey
	}
	err = kvs.writeType(typ)
	if err != nil {
		return 0, 0, err
	}

	if useCachedFlag {
		err = kvs.writer.writeLEU40(keyOffset)
		if err != nil {
			return 0, 0, err
		}
	} else {
		// Write key inline; for short cached keys this duplicates data but saves WAL space
		newOffset, err := kvs.writer.writeLenPrefixedBytes(unsafeKey)
		if err != nil {
			return 0, 0, err
		}
		if !cached {
			keyOffset = newOffset
			kvs.writer.addKeyToCache(key, keyOffset)
		}
	}

	valueOffset, err = kvs.writer.writeLenPrefixedBytes(value.UnsafeBytes())
	if err != nil {
		return 0, 0, err
	}

	return keyOffset, valueOffset, nil
}

// WriteWALDelete writes a WAL delete entry for the given key.
func (kvs *WALWriter) writeWALDelete(key UnsafeBytes) error {
	unsafeKey := key.UnsafeBytes()
	cachedOffsetAny, cached := kvs.writer.keyCache.Load(unsafeBytesToString(unsafeKey))
	var cachedOffset uint64
	if cached {
		cachedOffset = cachedOffsetAny.(uint64)
	}
	// Only use cached key flag if it saves space (offset is 5 bytes)
	useCachedFlag := cached && len(unsafeKey) >= 5
	typ := WALEntryDelete
	if useCachedFlag {
		typ |= WALFlagCachedKey
	}
	err := kvs.writeType(typ)
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
		keyOffset, err := kvs.writer.writeLenPrefixedBytes(unsafeKey)
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
func (kvs *WALWriter) writeWALCommit(version uint64, checkpoint bool) error {
	typ := WALEntryCommit
	if checkpoint {
		typ |= WALFlagCheckpoint
	}
	err := kvs.writeType(typ)
	if err != nil {
		return err
	}

	return kvs.writer.writeVarUint(version)
}

// Rollback rolls back the WAL writer to the state before the last WriteWALVersion call.
func (kvs *WALWriter) Rollback() error {
	currentSize := uint64(kvs.writer.Size())
	if kvs.lastVersionOffset == currentSize {
		// nothing to roll back
		return nil
	}
	if kvs.lastVersionOffset > currentSize {
		return fmt.Errorf("cannot rollback WAL writer: last version offset %d is not less than current size %d", kvs.lastVersionOffset, kvs.writer.Size())
	}

	// remove keys from the cache that were added in the current batch
	for _, update := range kvs.currentUpdates {
		var key []byte
		if update.SetNode != nil {
			key = update.SetNode.key
		} else {
			key = update.DeleteKey
		}

		if offset, found := kvs.writer.keyCache.Load(unsafeBytesToString(key)); found {
			if offset.(uint64) >= kvs.lastVersionOffset {
				kvs.writer.keyCache.Delete(unsafeBytesToString(key))
			}
		}
	}
	kvs.currentUpdates = nil

	// truncate the file back to the last version offset
	err := kvs.writer.file.Truncate(int64(kvs.lastVersionOffset))
	if err != nil {
		return fmt.Errorf("failed to truncate WAL file during rollback: %w", err)
	}
	_, err = kvs.writer.file.Seek(int64(kvs.lastVersionOffset), 0)
	if err != nil {
		return fmt.Errorf("failed to seek WAL file during rollback: %w", err)
	}

	// reset the writer
	kvs.writer.FileWriter = NewFileWriter(kvs.writer.file)
	kvs.writer.written = int(kvs.lastVersionOffset)

	if kvs.lastVersionOffset == 0 {
		// revert start version
		kvs.startVersion = 0
	}

	return nil
}

func (kvs *WALWriter) Sync() error {
	return kvs.writer.Sync()
}

func (kvs *WALWriter) Size() int {
	return kvs.writer.Size()
}

// LookupKeyOffset looks up the raw offset of the given key in the key cache.
func (kvs *WALWriter) LookupKeyOffset(key []byte) (uint64, bool) {
	offset, found := kvs.writer.keyCache.Load(unsafeBytesToString(key))
	if found {
		return offset.(uint64), true
	}
	return 0, false
}

func (kvs *WALWriter) writeType(x WALEntryType) error {
	_, err := kvs.writer.Write([]byte{byte(x)})
	return err
}
