package decoding

// SyncSource is an interface that allows indexers to start indexing modules with pre-existing state.
// It should generally be a wrapper around the key-value store.
type SyncSource interface {
	// IterateAllKVPairs iterates over all key-value pairs for a given module.
	IterateAllKVPairs(moduleName string, fn func(key, value []byte) error) error
}
