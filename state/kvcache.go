package state

// MemKVCache is designed to wrap MemKVStore as a cache
type MemKVCache struct {
	store      SimpleDB
	readCache  *MemKVStore
	writeCache *MemKVStore
}

var _ SimpleDB = (*MemKVCache)(nil)

// NewMemKVCache wraps a cache around MemKVStore
//
// You probably don't want to use directly, but rather
// via MemKVCache.Checkpoint()
func NewMemKVCache(store SimpleDB) *MemKVCache {
	if store == nil {
		panic("wtf")
	}

	return &MemKVCache{
		store:      store,
		readCache:  NewMemKVStore(),
		writeCache: NewMemKVStore(),
	}
}

// Set sets a key, fulfills KVStore interface
func (c *MemKVCache) Set(key []byte, value []byte) {
	// if we write it, remove from read cache
	c.readCache.Remove(key)
	c.writeCache.Set(key, value)
}

// Get gets a key, fulfills KVStore interface
func (c *MemKVCache) Get(key []byte) (value []byte) {
	// see if we have a cached write
	value, ok := c.writeCache.m[string(key)]
	if ok {
		return value
	}
	// then a cached read
	value, ok = c.readCache.m[string(key)]
	if ok {
		return value
	}
	// then read directly and cache it
	value = c.store.Get(key)
	c.readCache.Set(key, value)
	return value
}

// Has checks existence of a key, fulfills KVStore interface
func (c *MemKVCache) Has(key []byte) bool {
	value := c.Get(key)
	return value != nil
}

// Remove uses nil value as a flag to delete...
// not ideal but good enough for testing
func (c *MemKVCache) Remove(key []byte) (value []byte) {
	value = c.Get(key)
	c.Set(key, nil)
	return value
}

// List is also inefficiently implemented...
func (c *MemKVCache) List(start, end []byte, limit int) []Model {
	orig := c.store.List(start, end, 0)
	writen := c.writeCache.List(start, end, 0)
	keys := c.combineLists(orig, writen)

	// apply limit (too late)
	if limit > 0 && len(keys) > 0 {
		if limit > len(keys) {
			limit = len(keys)
		}
		keys = keys[:limit]
	}

	return keys
}

func (c *MemKVCache) combineLists(lists ...[]Model) []Model {
	store := NewMemKVStore()
	for _, cache := range lists {
		for _, m := range cache {
			if m.Value == nil {
				store.Remove([]byte(m.Key))
			} else {
				store.Set([]byte(m.Key), m.Value)
			}
		}
	}
	return store.List(nil, nil, 0)
}

// First is done with List, but could be much more efficient
func (c *MemKVCache) First(start, end []byte) Model {
	data := c.List(start, end, 0)
	if len(data) == 0 {
		return Model{}
	}
	return data[0]
}

// Last is done with List, but could be much more efficient
func (c *MemKVCache) Last(start, end []byte) Model {
	data := c.List(start, end, 0)
	if len(data) == 0 {
		return Model{}
	}
	return data[len(data)-1]
}

// Checkpoint returns the same state, but where writes
// are buffered and don't affect the parent
func (c *MemKVCache) Checkpoint() SimpleDB {
	return NewMemKVCache(c)
}

// Commit will take all changes from the checkpoint and write
// them to the parent.
// Returns an error if this is not a child of this one
func (c *MemKVCache) Commit(sub SimpleDB) error {
	cache, ok := sub.(*MemKVCache)
	if !ok {
		return ErrNotASubTransaction()
	}

	// see if it points to us
	ref, ok := cache.store.(*MemKVCache)
	if !ok || ref != c {
		return ErrNotASubTransaction()
	}

	// apply the cached data to us
	cache.applyCache()
	return nil
}

// applyCache will apply all the cache methods to the underlying store
func (c *MemKVCache) applyCache() {
	for _, k := range c.writeCache.keysInRange(nil, nil) {
		v := c.writeCache.m[k]
		if v == nil {
			c.store.Remove([]byte(k))
		} else {
			c.store.Set([]byte(k), v)
		}
	}
}

// Discard will remove reference to this
func (c *MemKVCache) Discard() {
	c.readCache = NewMemKVStore()
	c.writeCache = NewMemKVStore()
}
