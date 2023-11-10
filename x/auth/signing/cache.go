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

	cache, err := lru.New[string, []byte](500)
	if err != nil {
		panic(err)
	}
	defaultCache := &Cache{
		data: cache,
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

func (c *Cache) Get(key string) ([]byte, bool) {
	// validate
	if !c.validate(key) {
		return nil, false
	}

	return c.data.Get(key)
}

func (c *Cache) Add(key string, value []byte) {
	// validate
	if !c.validate(key) {
		return
	}
	// add cache
	c.data.Add(key, value)
}

func (c *Cache) Remove(key string) {
	// validate
	if !c.validate(key) {
		return
	}
	c.data.Remove(key)
}

func (c *Cache) validate(key string) bool {
	// validate key
	if len(key) == 0 {
		return false
	}

	return c.data != nil
}
