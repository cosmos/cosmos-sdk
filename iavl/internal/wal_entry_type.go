package internal

// WALEntryType represents the type of entry in the KV data file.
type WALEntryType byte

const (
	// WALEntryStart is the first entry in an uncompacted KV data file which indicates that this KV data file
	// can be used for WAL replay restoration. It must immediately be followed by the varint-encoded version number
	// corresponding to the first version in this changeset.
	WALEntryStart WALEntryType = 0x0

	// WALEntrySet indicates a set operation for a key-value pair.
	// This should be followed by:
	//   - varint key length + key bytes, OR if WALFlagCachedKey is set, a 32-bit LE offset to a cached key
	//   - varint value length + value bytes
	// Offsets point to the start of the varint length field, not the type byte.
	WALEntrySet WALEntryType = 0x1

	// WALEntryDelete indicates a delete operation for a key.
	// This should be followed by:
	//   - varint key length + key bytes, OR if WALFlagCachedKey is set, a 32-bit LE offset to a cached key
	// Offsets point to the start of the varint length field, not the type byte.
	WALEntryDelete WALEntryType = 0x2

	// WALEntryCommit indicates the commit operation for a version.
	// This must be followed by a varint-encoded version number.
	WALEntryCommit WALEntryType = 0x3

	// WALFlagCachedKey indicates that the key for this entry is cached and should be referenced by
	// a 32-bit little-endian offset instead of being stored inline.
	WALFlagCachedKey WALEntryType = 0x80

	// WALFlagCheckpoint modifies a commit entry to indicate that a checkpoint should have been saved for this commit.
	WALFlagCheckpoint WALEntryType = 0x40
)
