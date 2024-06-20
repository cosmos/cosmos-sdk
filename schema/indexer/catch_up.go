package indexer

// CatchUpSource is an interface that allows indexers to start indexing modules with pre-existing state.
type CatchUpSource interface {

	// IterateAllKVPairs iterates over all key-value pairs for a given module.
	IterateAllKVPairs(moduleName string, fn func(key, value []byte) error) error
}
