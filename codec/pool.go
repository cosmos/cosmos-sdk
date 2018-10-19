package codec

type CachePool struct {
	cdc  *Amino
	size int

	caches chan *cache
}

var _ Codec = (*CachePool)(nil)

func newCachePool(cdc *Amino, size int) *CachePool {
	return &CachePool{
		cdc:  cdc,
		size: size,

		caches: make(chan *cache),
	}
}

func (pool *CachePool) acquireCache() (res *cache) {
	select {
	case res = <-pool.caches:
		return
	default:
		res = &cache{
			cdc: newCache(pool.cdc, pool.size),
		}
		return
	}
}

func (pool *CachePool) returnCache(cache *cache) {
	pool.caches <- cache
}

func (pool *CachePool) MarshalJSON(o interface{}) (res []byte, err error) {
	cache := pool.acquireCache()
	res, err = cache.MarshalJSON(o)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) UnmarshalJSON(bz []byte, ptr interface{}) (err error) {
	cache := pool.acquireCache()
	err = cache.UnmarshalJSON(bz, ptr)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) MarshalBinary(o interface{}) (res []byte, err error) {
	cache := pool.acquireCache()
	res, err = cache.MarshalBinary(o)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) UnmarshalBinary(bz []byte, ptr interface{}) (err error) {
	cache := pool.acquireCache()
	err = cache.UnmarshalBinary(bz, ptr)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) MustMarshalBinary(o interface{}) (res []byte) {
	cache := pool.acquireCache()
	res = cache.MustMarshalBinary(o)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) MustUnmarshalBinary(bz []byte, ptr interface{}) {
	cache := pool.acquireCache()
	cache.MustUnmarshalBinary(bz, ptr)
	pool.returnCache(cache)
	return
}
