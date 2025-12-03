package internal

// KVEntryType represents the type of entry in the KV data file.
type KVEntryType byte

const (
	// KVEntryWALStart is the first entry in an uncompacted KV data file which indicates that this KV data file
	// can be used for WAL replay restoration. It must immediately be followed by the varint-encoded version number
	// corresponding to the first version in this changeset.
	KVEntryWALStart KVEntryType = 0x0

	// KVEntryWALSet indicates a set operation for a key-value pair.
	// This should be followed by a varint-encoded length and the raw bytes OR
	// if the KVFlagCachedKey flag is set, a 32-bit little-endian offset referencing a cached key,
	// AND then a varint-encoded length and the value bytes.
	KVEntryWALSet KVEntryType = 0x1

	// KVEntryWALDelete indicates a delete operation for a key.
	// This should be followed by a varint-encoded length and the key bytes OR
	// if the KVFlagCachedKey flag is set, a 32-bit little-endian offset referencing a cached key.
	KVEntryWALDelete KVEntryType = 0x2

	// KVEntryWALCommit indicates the commit operation for a version.
	// This must be followed by a varint-encoded version number.
	KVEntryWALCommit KVEntryType = 0x3

	// KVEntryBlob indicates a blob entry storing raw key-value data.
	// This should be followed by a varint-encoded length and the raw bytes.
	// This entry type is used for compacted (non-WAL) KV data files or
	// for branch keys that aren't otherwise cached.
	KVEntryBlob KVEntryType = 0x4

	// KVFlagCachedKey indicates that the key for this entry is cached and should be referenced by
	// a 32-bit little-endian offset instead of being stored inline.
	KVFlagCachedKey KVEntryType = 0x80
)
