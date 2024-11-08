package branch

type memCache struct {
	items map[string][]byte
}

func newMemCache() memCache {
	return memCache{items: make(map[string][]byte)}
}

func (m memCache) get(key []byte) ([]byte, bool) {
	v, ok := m.items[unsafeString(key)]
	return v, ok
}

func (m memCache) set(key, value []byte) {
	// we do not use unsafe because these are stored
	// indefinitely in the cache
	m.items[string(key)] = value
}

func (m memCache) delete(key []byte) {
	m.items[string(key)] = nil
}
