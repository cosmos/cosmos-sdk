package signing

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

var signatureCache *Cache

const (
	TxHashLen        = 32
	AddressStringLen = 2 + 20*2
)

var initSigCacheOnce sync.Once

// Cache is a cache of verified signatures
type Cache struct {
	data *lru.Cache[string, []byte]
}

// NewSignatureCache initializes the signature cache.
// signature verification is one of the most expensive parts in verification
// by caching it we avoid needing to verify the same signature multiple times
// This is only initialized once.
func NewSignatureCache() *Cache {
	sc := &Cache{}
	initSigCacheOnce.Do(func() {
		// 500 * (32 + 42) = 37.5KB
		cache, err := lru.New[string, []byte](500)
		if err != nil {
			panic(err)
		}
		sc = &Cache{data: cache}
	})

	return sc
}

func SignatureCache() *Cache {
	return signatureCache
}

// Get returns the cached signature if it exists
func (c *Cache) Get(key string) ([]byte, bool) {
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

// sigkey is the key used to store the signature in the cache
type sigkey struct {
	signbytes []byte
	sig       []byte
}

func newSigKey(signbytes, sig []byte) sigkey {
	return sigkey{
		signbytes: signbytes,
		sig:       sig,
	}
}

func (s sigkey) string() string {
	return string(append(s.signbytes, s.sig...))
}
