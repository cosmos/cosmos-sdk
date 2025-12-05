package internal

// KVEntryType represents the type of entry in the KV data file.
type KVEntryType byte

const (
	// KVEntryWALStart is the first entry in an uncompacted KV data file which indicates that this KV data file
	// can be used for WAL replay restoration. It must immediately be followed by the varint-encoded version number
	// corresponding to the first version in this changeset.
	KVEntryWALStart KVEntryType = 0x0

	// KVEntryWALSet indicates a set operation for a key-value pair.
	// This should be followed by:
	//   - varint key length + key bytes, OR if KVFlagCachedKey is set, a 32-bit LE offset to a cached key
	//   - varint value length + value bytes
	// Offsets point to the start of the varint length field, not the type byte.
	KVEntryWALSet KVEntryType = 0x1

	// KVEntryWALDelete indicates a delete operation for a key.
	// This should be followed by:
	//   - varint key length + key bytes, OR if KVFlagCachedKey is set, a 32-bit LE offset to a cached key
	// Offsets point to the start of the varint length field, not the type byte.
	KVEntryWALDelete KVEntryType = 0x2

	// KVEntryWALCommit indicates the commit operation for a version.
	// This must be followed by a varint-encoded version number.
	KVEntryWALCommit KVEntryType = 0x3

	// KVEntryKeyBlob indicates a standalone key data entry.
	// This should be followed by varint length + raw bytes.
	// Used for compacted (non-WAL) leaf or branch keys not already cached.
	KVEntryKeyBlob KVEntryType = 0x4

	// KVEntryValueBlob indicates a standalone value data entry.
	// This should be followed by varint length + raw bytes.
	// Used for compacted (non-WAL) leaf values.
	// The main difference between KVEntryKeyBlob and KVEntryValueBlob is that key
	// entries may be cached for faster access, while value entries are not cached.
	KVEntryValueBlob KVEntryType = 0x5

	// KVFlagCachedKey indicates that the key for this entry is cached and should be referenced by
	// a 32-bit little-endian offset instead of being stored inline.
	KVFlagCachedKey KVEntryType = 0x80
)
