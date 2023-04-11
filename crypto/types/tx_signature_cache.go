package types

import (
	"sync/atomic"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/spf13/viper"
)

var (
	signatureCache *Cache
)

const (
	TxHashLen        = 32
	AddressStringLen = 2 + 20*2
)

const FlagSigCacheSize = "signature-cache-size"

func init() {
	// used for ut
	defaultCache := &Cache{
		data:      nil,
		readCount: 0,
		hitCount:  0,
	}
	signatureCache = defaultCache
}

func InitSignatureCache() {
	fastCache := fastcache.New((TxHashLen + AddressStringLen) * viper.GetInt(FlagSigCacheSize))
	signatureCache = &Cache{
		data: fastCache,
	}
}

func SignatureCache() *Cache {
	return signatureCache
}

type Cache struct {
	data      *fastcache.Cache
	readCount int64
	hitCount  int64
}

func (c *Cache) Get(key []byte) ([]byte, bool) {
	// validate
	if !c.validate(key) {
		return nil, false
	}
	atomic.AddInt64(&c.readCount, 1)
	// get cache
	value, ok := c.data.HasGet(nil, key)
	if ok {
		atomic.AddInt64(&c.hitCount, 1)
		return value, true
	}
	return nil, false
}

func (c *Cache) Add(key []byte, value []byte) {
	// validate
	if !c.validate(key) {
		return
	}
	// add cache
	c.data.Set(key, value)
}

func (c *Cache) Remove(key []byte) {
	// validate
	if !c.validate(key) {
		return
	}
	c.data.Del(key)
}

func (c *Cache) ReadCount() int64 {
	return atomic.LoadInt64(&c.readCount)
}

func (c *Cache) HitCount() int64 {
	return atomic.LoadInt64(&c.hitCount)
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
