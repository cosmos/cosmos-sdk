package internal

import (
	"context"
	"fmt"
	"os"
)

// WALWriter writes the write-ahead log (WAL) (wal.log) of updates to the tree.
// A WAL consists of:
// - a start version entry with the first version in the WAL
// - sets or deletes for a version
// - followed by a commit entry for each committed version
type WALWriter struct {
	// writer is the underlying writer that handles the low-level details of writing WAL entries and managing the key cache.
	// This type is also used for writing other non-WAL key-value data that needs to be written to kv.data.
	writer            *KVDataWriter
	startVersion      uint64
	lastVersionOffset uint64
	// currentUpdates holds the current batch of updates that have just been written,
	// in case we need to roll back
	currentUpdates []KVUpdate
}

// NewWALWriter creates a new WALWriter for the provided file.
func NewWALWriter(file *os.File) *WALWriter {
	return &WALWriter{
		writer: NewKVDataWriter(file),
	}
}

// WALWriteAbortedErr is returned when the WAL file was successfully rolled back due to another error when writing
// a WAL version.
// This error will usually wrap another error, so errors.Is() should be used.
var WALWriteAbortedErr = fmt.Errorf("WAL write aborted and rolled back")

// WriteWALVersion writes a batch of updates at once.
// This should be the only method called externally to write the WAL as this method ensures that
// a proper state is maintained to support rollbacks.
// If there is an error when writing the WAL, a rollback will be attempted.
// In this case if the rollback was successful,
// the error will be wrapped in a WALWriteAbortedErr to indicate that the WAL was rolled back due to this error.
// Use errors.Is() to test this.
// If there was an error and then an error rolling back, then the error will not be wrapped in WALWriteAbortedErr.
// If the context passed in is cancelled at any point during writing the WAL, a rollback will be attempted automatically.
func (kvs *WALWriter) WriteWALVersion(ctx context.Context, version uint64, updates []KVUpdate, checkpoint bool) error {
	// Keep track of the current file offset so that we can rollback to it if needed.
	kvs.lastVersionOffset = uint64(kvs.writer.Size())
	// Keep track of this batch of updates so that we can leave the key cache in a clean state after a rollback.
	kvs.currentUpdates = updates

	if err := kvs.doWriteWALVersion(ctx, version, updates, checkpoint); err != nil {
		// Rollback if we failed for some reason.
		rbErr := kvs.Rollback()
		if rbErr != nil {
			return fmt.Errorf("failed to write WAL version: %w; rollback also failed: %w", err, rbErr)
		}
		return fmt.Errorf("%w; due to: %w", WALWriteAbortedErr, err)
	}
	return nil
}

// doWriteWALVersion does the mechanical writing of a WAL version but doesn't deal with rollback as WriteWALVersion above does.
func (kvs *WALWriter) doWriteWALVersion(ctx context.Context, version uint64, updates []KVUpdate, checkpoint bool) error {
	// Write the start version entry only if this is the first version in this WAL file.
	if err := kvs.writeStartVersion(version); err != nil {
		return err
	}

	// Write the updates sequentially, we will check for context cancellation before each update.
	if err := kvs.writeWALUpdates(ctx, updates); err != nil {
		return err
	}

	// Check for context cancellation before committing to rollback fast.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Write the commit entry for this version.
	return kvs.writeWALCommit(version, checkpoint)
}

// writeStartVersion writes a WAL start entry if the start version is not already set.
func (kvs *WALWriter) writeStartVersion(version uint64) error {
	// If startVersion is not 0 this means we have already written a start version entry to this WAL file
	// and we have nothing to do.
	if kvs.startVersion != 0 {
		// start version already set
		return nil
	}
	// Write the start version and keep track of which version we started at.
	kvs.startVersion = version
	err := kvs.writeType(WALEntryStart)
	if err != nil {
		return err
	}
	return kvs.writer.writeVarUint(version)
}

// WriteWALUpdates writes a batch of WAL updates.
func (kvs *WALWriter) writeWALUpdates(ctx context.Context, updates []KVUpdate) error {
	for _, update := range updates {
		// Check for context cancellation before writing every update so we can fail fast
		// and avoid writing partial updates to the WAL if the context is cancelled.
		if err := ctx.Err(); err != nil {
			return err
		}

		// Every update must be either a delete or a set, but not both
		deleteKey := update.DeleteKey
		setNode := update.SetNode
		if deleteKey != nil && setNode != nil {
			return fmt.Errorf("invalid update: both SetNode and DeleteKey are set")
		}

		if deleteKey == nil && setNode == nil {
			return fmt.Errorf("invalid update: neither SetNode nor DeleteKey is set")
		}

		if deleteKey != nil {
			// Write the WAL delete operation to disk
			err := kvs.writeWALDelete(WrapSafeBytes(deleteKey))
			if err != nil {
				return err
			}
		} else { // setNode != nil
			// Write the WAL set operation to disk
			keyOffset, valueOffset, err := kvs.writeWALSet(WrapSafeBytes(setNode.key), WrapSafeBytes(setNode.value))
			if err != nil {
				return err
			}
			// setNode is actually a leaf MemNode in the root of the new tree.
			// We store the key and value offsets in this MemNode so that layer when
			// we persist this leaf node to disk as a LeafLayout, we know the offsets in the WAL file
			setNode.walKeyOffset = keyOffset
			setNode.walValueOffset = valueOffset
		}
	}
	return nil
}

// WriteWALSet writes a WAL set entry for the given key and value and returns their raw offsets.
func (kvs *WALWriter) writeWALSet(key, value UnsafeBytes) (keyOffset, valueOffset uint64, err error) {
	keyOffset, err = kvs.writeWALKey(key, WALEntrySet)
	if err != nil {
		return 0, 0, err
	}

	valueOffset, err = kvs.writer.writeLenPrefixedBytes(value.UnsafeBytes())
	if err != nil {
		return 0, 0, err
	}

	return keyOffset, valueOffset, nil
}

// WriteWALDelete writes a WAL delete entry for the given key.
func (kvs *WALWriter) writeWALDelete(key UnsafeBytes) error {
	_, err := kvs.writeWALKey(key, WALEntryDelete)
	return err
}

// writeWALKey writes the key entry for a WAL set or delete operation, and returns the offset of the key in the WAL file.
// If the key was already written to the disk and is >= 5 bytes long, we will find it in the
// key cache and then write a pointer to the offset at which it was previously written instead of writing the full entry.
// This key deduplication is intended to save space on disk when the same keys are frequently updated and should
// only cost a small amount of memory.
func (kvs *WALWriter) writeWALKey(key UnsafeBytes, entryType WALEntryType) (keyOffset uint64, err error) {
	unsafeKey := key.UnsafeBytes()
	// safe to use the unsafe key bytes for lookup
	keyOffsetAny, cached := kvs.writer.keyCache.Load(unsafeBytesToString(unsafeKey))
	if cached {
		keyOffset = keyOffsetAny.(uint64)
	}
	// We only write cached entries where the key is longer than 5 bytes because the offset is 5 bytes (40 bits).
	useCachedFlag := cached && len(unsafeKey) >= 5
	if useCachedFlag {
		// If we are writing a cached key, we set the WALFlagCachedKey flag and write the offset instead of the key bytes.
		entryType |= WALFlagCachedKey
	}
	err = kvs.writeType(entryType)
	if err != nil {
		return 0, err
	}

	if useCachedFlag {
		// Write the offset instead of key bytes.
		err = kvs.writer.writeLEU40(keyOffset)
		if err != nil {
			return 0, err
		}
	} else {
		// Write key inline; for short cached keys this duplicates data but saves WAL space.
		newOffset, err := kvs.writer.writeLenPrefixedBytes(unsafeKey)
		if err != nil {
			return 0, err
		}
		if !cached {
			keyOffset = newOffset
			kvs.writer.addKeyToCache(key, keyOffset)
		}
	}

	return keyOffset, nil
}

// WriteWALCommit writes a WAL commit entry for the given version with a flag to indicate whether a checkpoint
// should have been written.
// The checkpoint flag could allow us to recreate failed checkpoints at the right height in the future.
// For now, if a checkpoint fails, we just roll it back and try again at a later height.
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
// This involves:
// - truncating the WAL file back to its previous height
// - making sure that we leave the key cache in a clean state and do not leave any entries in the cache that reference
//   data in the truncated/rolled-back region
func (kvs *WALWriter) Rollback() error {
	currentSize := uint64(kvs.writer.Size())
	if kvs.lastVersionOffset == currentSize {
		// nothing to roll back
		return nil
	}
	if kvs.lastVersionOffset > currentSize {
		return fmt.Errorf("cannot rollback WAL writer: last version offset %d is not less than current size %d", kvs.lastVersionOffset, kvs.writer.Size())
	}

	// Remove keys from the cache that were added in the current batch of updates.
	// If we leave these entries in the cache, they will point to offsets in the WAL file that have been rolled back and
	// may be overwritten by future writes, which could lead to incorrect values being read from the WAL file.
	for _, update := range kvs.currentUpdates {
		var key []byte
		if update.SetNode != nil {
			key = update.SetNode.key
		} else {
			key = update.DeleteKey
		}

		if offset, found := kvs.writer.keyCache.Load(unsafeBytesToString(key)); found {
			if offset.(uint64) >= kvs.lastVersionOffset {
				// If we have an offset that is past the rollback offset, then we must clear it from the cache.
				kvs.writer.keyCache.Delete(unsafeBytesToString(key))
			}
		}
	}
	// Clear the list of current updates.
	kvs.currentUpdates = nil

	// Truncate the file back to the last version offset.
	err := kvs.writer.file.Truncate(int64(kvs.lastVersionOffset))
	if err != nil {
		return fmt.Errorf("failed to truncate WAL file during rollback: %w", err)
	}
	// Make sure that the file handle is now at the last version offset.
	_, err = kvs.writer.file.Seek(int64(kvs.lastVersionOffset), 0)
	if err != nil {
		return fmt.Errorf("failed to seek WAL file during rollback: %w", err)
	}

	// Reset the writer to clear any buffered data.
	kvs.writer.FileWriter = NewFileWriter(kvs.writer.file)
	// Make sure that the file writer is initialized with the correct size (note that the file writer should have the write size by default now, but this doesn't hurt).
	kvs.writer.written = int(kvs.lastVersionOffset)

	if kvs.lastVersionOffset == 0 {
		// If we rolled the WAL all the way back to the start of the file,
		// we must also reset startVersion so that when we start writing this file again,
		// we will emit a start version entry to properly initialize the WAL.
		// Otherwise, the WAL will be corrupted.
		kvs.startVersion = 0
	}

	return nil
}

// Sync fsync's the WAL file to disk.
func (kvs *WALWriter) Sync() error {
	return kvs.writer.Sync()
}

// Size returns the current size of the WAL file.
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

// writeType writes the single byte type entry for a WAL entry.
func (kvs *WALWriter) writeType(x WALEntryType) error {
	_, err := kvs.writer.Write([]byte{byte(x)})
	return err
}
