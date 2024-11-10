package object

import (
	"unsafe"

	corestore "cosmossdk.io/core/store"
)

var _ corestore.ObjectStore = (*Cache)(nil)

// Cache is a map that stores objects in any
type Cache struct {
	m map[string]any
}

// NewCache creates a new cache
func NewCache() *Cache {
	return &Cache{
		m: make(map[string]any),
	}
}

// Get returns nil if key doesn't exist. Panics on nil key.
// Contract: The set value must removed at the end of the block.
func (c Cache) Set(prefix []byte, value any) {
	_, exists := c.m[unsafeString(prefix)]
	if exists {
		c.m[unsafeString(prefix)] = value
	}
	c.m[unsafeString(prefix)] = value
}

// Get returns nil if key doesn't exist.
func (c Cache) Get(prefix []byte) (value any, ok bool) {
	value, ok = c.m[unsafeString(prefix)]
	return
}

// Delete deletes the key.
func (c Cache) Delete(prefix []byte) {
	delete(c.m, unsafeString(prefix))
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
