package signing

import (
	lru "github.com/hashicorp/golang-lru/v2"
)

var signatureCache *Cache

const (
	TxHashLen        = 32
	AddressStringLen = 2 + 20*2
)

func init() {
	// used for ut
	defaultCache := &Cache{
		data: &lru.Cache[string, []byte]{},
	}
	signatureCache = defaultCache
}

// InitSignatureCache initializes the signature cache.
// signature verification is one of the most expensive parts in verification
// by caching it we avoid needing to verify the same signature multiple times
func NewSignatureCache() {
	// 500 * (32 + 42) = 37.5KB
	cache, err := lru.New[string, []byte](500)
	if err != nil {
		panic(err)
	}
	signatureCache = &Cache{
		data: cache,
	}
}

func SignatureCache() *Cache {
	return signatureCache
}

type Cache struct {
	data *lru.Cache[string, []byte]
}

func (c *Cache) Get(key []byte) ([]byte, bool) {
	// validate
	if !c.validate(key) {
		return nil, false
	}
	// get cache
	value, ok := c.data.Get(string(key))
	if ok {
		return value, true
	}
	return nil, false
}

func (c *Cache) Add(key, value []byte) {
	// validate
	if !c.validate(key) {
		return
	}
	// add cache
	c.data.Add(string(key), value)
}

func (c *Cache) Remove(key []byte) {
	// validate
	if !c.validate(key) {
		return
	}
	c.data.Remove(string(key))
}

func (c *Cache) validate(key []byte) bool {
	// validate key
	if len(key) == 0 {
		return false
	}
	// validate lru cache
	if c.data == nil {
		return false
	}
	return true
}
