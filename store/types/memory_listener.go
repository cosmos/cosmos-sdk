package types

// MemoryListener listens to the state writes and accumulate the records in memory.
type MemoryListener struct {
	stateCache []*StoreKVPair
}

func NewMemoryListener() *MemoryListener {
	return &MemoryListener{}
}

// OnWrite records a state write in memory.
func (fl *MemoryListener) OnWrite(storeKey string, key, value []byte, delete bool) {
	fl.stateCache = append(fl.stateCache, &StoreKVPair{
		StoreKey: storeKey,
		Delete:   delete,
		Key:      key,
		Value:    value,
	})
}

// PopStateCache returns the current state caches and set to nil.
func (fl *MemoryListener) PopStateCache() []*StoreKVPair {
	res := fl.stateCache
	fl.stateCache = nil
	return res
}
