package state

// MemKVCache is designed to wrap MemKVStore as a cache
type MemKVCache struct {
	store SimpleDB
	cache *MemKVStore
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
		store: store,
		cache: NewMemKVStore(),
	}
}

func (c *MemKVCache) Set(key []byte, value []byte) {
	c.cache.Set(key, value)
}

func (c *MemKVCache) Get(key []byte) (value []byte) {
	value, ok := c.cache.m[string(key)]
	if !ok {
		value = c.store.Get(key)
		c.cache.Set(key, value)
	}
	return value
}

func (c *MemKVCache) Has(key []byte) bool {
	value := c.Get(key)
	return value != nil
}

// Remove uses nil value as a flag to delete... not ideal but good enough
// for testing
func (c *MemKVCache) Remove(key []byte) (value []byte) {
	value = c.Get(key)
	c.cache.Set(key, nil)
	return value
}

// List is also inefficiently implemented...
func (c *MemKVCache) List(start, end []byte, limit int) []Model {
	orig := c.store.List(start, end, 0)
	cached := c.cache.List(start, end, 0)
	keys := c.combineLists(orig, cached)

	// apply limit (too late)
	if limit > 0 && len(keys) > 0 {
		if limit > len(keys) {
			limit = len(keys)
		}
		keys = keys[:limit]
	}

	return keys
}

func (c *MemKVCache) combineLists(orig, cache []Model) []Model {
	store := NewMemKVStore()
	for _, m := range orig {
		store.Set(m.Key, m.Value)
	}
	for _, m := range cache {
		if m.Value == nil {
			store.Remove([]byte(m.Key))
		} else {
			store.Set([]byte(m.Key), m.Value)
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
	// TODO: see if it points to us

	// apply the cached data to us
	cache.applyCache()
	return nil
}

// applyCache will apply all the cache methods to the underlying store
func (c *MemKVCache) applyCache() {
	for k, v := range c.cache.m {
		if v == nil {
			c.store.Remove([]byte(k))
		} else {
			c.store.Set([]byte(k), v)
		}
	}
}

// Discard will remove reference to this
func (c *MemKVCache) Discard() {
	c.cache = NewMemKVStore()
}
