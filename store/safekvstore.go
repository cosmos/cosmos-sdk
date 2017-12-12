package store

// SafeKVStore is a KVStore that includes a WriteLocker to
// make caches thread-safe to write.
type SafeKVStore struct {
	KVStore
	vrwMutex
}

func NewSafeKVStore(st KVStore) *SafeKVStore {
	return &SafeKVStore{
		KVStore:  st,
		vrwMutex: NewVRWMutex(),
	}
}

func (st *SafeKVStore) Get(key []byte) []byte {
	st.RLock()
	defer st.RUnlock()

	return st.KVStore.Get(key)
}

func (st *SafeKVStore) Has(key []byte) bool {
	st.RLock()
	defer st.RUnlock()

	return st.KVStore.Has(key)
}

func (st *SafeKVStore) Set(key, value []byte) {
	st.Lock()
	defer st.Unlock()

	return st.KVStore.Set(key, value)
}

func (st *SafeKVStore) Delete(key) {
	st.Lock()
	defer st.Unlock()

	return st.KVStore.Delete(key)
}

func (st *SafeKVStore) Iterator(start, end []byte) {
	st.RLock()
	defer st.RUnlock()

	return st.KVStore.Iterator(start, end)
}
