package internal

// KVEntryType represents the type of entry in the KV data file.
type KVEntryType byte

const (
	// KVEntryWALStart is the first entry in an uncompacted KV data file which indicates that this KV data file
	// can be used for WAL replay restoration. It must immediately be followed by the varint-encoded version number
	// corresponding to the first version in this changeset.
	KVEntryWALStart KVEntryType = 0x0

	// KVEntryWALSet indicates a set operation for a key-value pair.
	// This must be followed by:
	// - 16-bit little-endian length-prefixed key or 32-bit offset to cached key if KVFlagCachedKey is set
	// - varint length-prefixed value
	KVEntryWALSet KVEntryType = 0x1

	// KVEntryWALDelete indicates a delete operation for a key.
	// This must be followed by:
	// - 16-bit little-endian length-prefixed key or 32-bit offset to cached key if KVFlagCachedKey is set
	KVEntryWALDelete KVEntryType = 0x2

	// KVEntryWALCommit indicates the end of a batch of operations for a specific version.
	// This must be followed by a varint-encoded version number.
	KVEntryWALCommit KVEntryType = 0x3

	// KVEntryKeyData indicates a key entry in the KV data file.
	// This must be followed by:
	// - 16-bit little-endian length-prefixed key
	KVEntryKeyData KVEntryType = 0x4

	// KVEntryKeyValueData indicates a value entry in the KV data file.
	// This must be followed by:
	// - 16-bit little-endian length-prefixed key or 32-bit offset to cached key if KVFlagCachedKey is set
	// - varint length-prefixed value
	KVEntryKeyValueData KVEntryType = 0x5

	// KVFlagCachedKey indicates that the key is stored as a 32-bit offset to a cached key entry.
	KVFlagCachedKey KVEntryType = 0x80
)
