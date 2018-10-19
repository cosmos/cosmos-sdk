package codec

type CachePool struct {
	cdc  *Amino
	size int

	caches chan *cache
}

var _ Codec = (*CachePool)(nil)

const maxcache = 5

func newCachePool(cdc *Amino, size int) *CachePool {
	return &CachePool{
		cdc:  cdc,
		size: size,

		caches: make(chan *cache, maxcache),
	}
}

func (pool *CachePool) acquireCache() (res *cache) {
	if len(pool.caches) >= maxcache {
		res = <-pool.caches
		return
	}

	select {
	case res = <-pool.caches:
		return
	default:
		res = newCache(pool.cdc, pool.size)
		return
	}
}

func (pool *CachePool) returnCache(cache *cache) {
	pool.caches <- cache
}

func (pool *CachePool) MarshalJSON(o interface{}) ([]byte, error) {
	return pool.cdc.MarshalJSON(o)
}

func (pool *CachePool) UnmarshalJSON(bz []byte, ptr interface{}) (err error) {
	cache := pool.acquireCache()
	err = cache.UnmarshalJSON(bz, ptr)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) MarshalBinary(o interface{}) ([]byte, error) {
	return pool.cdc.MarshalBinary(o)
}

func (pool *CachePool) UnmarshalBinary(bz []byte, ptr interface{}) (err error) {
	cache := pool.acquireCache()
	err = cache.UnmarshalBinary(bz, ptr)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) MustMarshalBinary(o interface{}) []byte {
	return pool.cdc.MustMarshalBinary(o)
}

func (pool *CachePool) MustUnmarshalBinary(bz []byte, ptr interface{}) {
	cache := pool.acquireCache()
	cache.MustUnmarshalBinary(bz, ptr)
	pool.returnCache(cache)
	return
}
