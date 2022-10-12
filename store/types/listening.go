package types

// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// OnWrite captures store state changes
	// if value is nil then it was deleted
	// storeKey indicates the source KVStore, to facilitate using the same WriteListener across separate KVStores
	// delete bool indicates if it was a delete; true: delete, false: set
	OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool)
}

// MemoryListener listens to the state writes and accumulate the records in memory.
type MemoryListener struct {
	key        StoreKey
	stateCache []StoreKVPair
}

// NewMemoryListener creates a listener that accumulate the state writes in memory.
func NewMemoryListener(key StoreKey) *MemoryListener {
	return &MemoryListener{key: key}
}

// OnWrite implements WriteListener interface
func (fl *MemoryListener) OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool) error {
	fl.stateCache = append(fl.stateCache, StoreKVPair{
		StoreKey: storeKey.Name(),
		Delete:   delete,
		Key:      key,
		Value:    value,
	})
	return nil
}

// PopStateCache returns the current state caches and set to nil
func (fl *MemoryListener) PopStateCache() []StoreKVPair {
	res := fl.stateCache
	fl.stateCache = nil
	return res
}

// StoreKey returns the storeKey it listens to
func (fl *MemoryListener) StoreKey() StoreKey {
	return fl.key
}
