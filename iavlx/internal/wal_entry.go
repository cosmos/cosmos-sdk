package internal

// WALOpType identifies the type of a WAL entry.
type WALOpType int

const (
	// WALOpSet records that a key-value pair was set.
	WALOpSet WALOpType = iota
	// WALOpDelete records that a key was deleted.
	WALOpDelete
	// WALOpCommit marks the end of a version's updates (the commit boundary).
	// A version's WAL entries are: zero or more Set/Delete entries followed by exactly one Commit.
	WALOpCommit
)

// WALEntry is a single entry read from the WAL during replay.
type WALEntry struct {
	// Op is the operation type (set, delete, or commit).
	Op WALOpType
	// Version is the tree version this entry belongs to.
	Version uint64
	// Key is the key for Set/Delete operations. Backed by the WAL's mmap — must be SafeCopy'd
	// if it needs to outlive the mmap pin.
	Key UnsafeBytes
	// Value is the value for Set operations (empty for Delete/Commit). Same mmap caveat as Key.
	Value UnsafeBytes
	// KeyOffset is the byte offset of the key blob in the WAL file. Used by checkpoint writing
	// to reference WAL data by offset instead of copying it (see WALWriter.LookupKeyOffset).
	KeyOffset int
	// ValueOffset is the byte offset of the value blob in the WAL file.
	ValueOffset int
	// Checkpoint is true if this commit entry also triggers a checkpoint.
	Checkpoint bool
	// EndOffset is the file offset immediately after the end of this WAL entry, used for rollback
	// (we truncate the file to this offset to undo everything after this entry).
	EndOffset int
}
