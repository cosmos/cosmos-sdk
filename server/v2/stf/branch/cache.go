package branch

import "unsafe"

func newMemoryCache() memoryCache {
	return memoryCache{
		items: map[string][]byte{},
	}
}

type memoryCache struct {
	items map[string][]byte
}

func (l memoryCache) get(key []byte) (value []byte, found bool) {
	keyStr := unsafe.String(unsafe.SliceData(key), len(key))

	value, found = l.items[keyStr]
	return
}

func (l memoryCache) set(key, value []byte) {
	l.items[string(key)] = value
}

func (l memoryCache) delete(key []byte) {
	l.items[string(key)] = nil
}
