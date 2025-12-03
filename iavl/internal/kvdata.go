package internal

// KVDataEntryType represents the type of entry in the KV data file.
type KVDataEntryType byte

const (
	// KVDataEntryTypeWALStart is the first entry in an uncompacted KV data file which indicates that this KV data file
	// can be used for WAL replay restoration. It must immediately be followed by the varint-encoded version number
	// corresponding to the first version in this changeset.
	KVDataEntryTypeWALStart KVDataEntryType = iota
	// KVDataEntryTypeWALSet indicates a set operation for a key-value pair.
	KVDataEntryTypeWALSet = iota
	KVDataEntryTypeWALDelete
	KVDataEntryTypeWALCommit
	KVDataEntryTypeExtraK
	KVDataEntryTypeExtraKV
)
