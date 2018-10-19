package codec

import (
	"math/rand"
	"testing"
)

func TestCachePoolPoolIdentity(t *testing.T) {
	size := 100

	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	cache := cdc.Cache(size)

	datanum := size * 10
	tcs := randgen(datanum)
	for i := 0; i < datanum*10; i++ {
		index := int(rand.ExpFloat64()) * datanum % datanum
		tc := tcs[index]

		testEqual(t, tc, cdc.MarshalJSON, cdc.UnmarshalJSON, cache.MarshalJSON, cache.UnmarshalJSON)
		testEqual(t, tc, cdc.MarshalBinary, cdc.UnmarshalBinary, cache.MarshalBinary, cache.UnmarshalBinary)
		testEqualMust(t, tc, cdc.MustMarshalBinary, cdc.MustUnmarshalBinary, cache.MustMarshalBinary, cache.MustUnmarshalBinary)
	}
}

func benchmarkCachePool(b *testing.B, cdc *CachePool, datanum, lambda int) {
	tcs := randgen(datanum)
	b.ResetTimer()

	var chit int
	for i := 0; i < b.N; i++ {

		index := int(rand.ExpFloat64()*float64(datanum)/float64(lambda)) % datanum
		tc := tcs[index]
		ptr := tc.ptr()

		bz, _ := cdc.MarshalJSON(tc.value)
		hit, _ := cdc.logUnmarshalJSON(bz, ptr)
		chit += hit

		bz, _ = cdc.MarshalBinary(tc.value)
		hit, _ = cdc.logUnmarshalBinary(bz, ptr)
		chit += hit
	}
	b.Logf("Cache hit %.2f%%", float64(chit)/2/float64(b.N)*100)
}

// Size 1000 Lambda 10
func BenchmarkNoCachePoolSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 1000, 10)
}
func BenchmarkP2PercentCachePoolSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(2), 1000, 10)
}
func BenchmarkP5PercentCachePoolSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(5), 1000, 10)
}
func Benchmark1PercentCachePoolSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(10), 1000, 10)
}
func Benchmark2PercentCachePoolSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(20), 1000, 10)
}
func Benchmark5PercentCachePoolSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(50), 1000, 10)
}

// Size 100000 Lambda 10
func BenchmarkNoCachePoolSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 100000, 10)
}
func BenchmarkP2PercentCachePoolSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(200), 100000, 10)
}
func BenchmarkP5PercentCachePoolSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(500), 100000, 10)
}
func Benchmark1PercentCachePoolSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(1000), 100000, 10)
}
func Benchmark2PercentCachePoolSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(2000), 100000, 10)
}
func Benchmark5PercentCachePoolSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(5000), 100000, 10)
}

// Size 1000 Lambda 50
func BenchmarkNoCachePoolSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 1000, 50)
}
func BenchmarkP2PercentCachePoolSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(2), 1000, 50)
}
func BenchmarkP5PercentCachePoolSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(5), 1000, 50)
}
func Benchmark1PercentCachePoolSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(10), 1000, 50)
}
func Benchmark2PercentCachePoolSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(20), 1000, 50)
}
func Benchmark5PercentCachePoolSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(50), 1000, 50)
}

// Size 100000 Lambda 50
func BenchmarkNoCachePoolSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 100000, 50)
}
func BenchmarkP2PercentCachePoolSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(200), 100000, 50)
}
func BenchmarkP5PercentCachePoolSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(500), 100000, 50)
}
func Benchmark1PercentCachePoolSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(1000), 100000, 50)
}
func Benchmark2PercentCachePoolSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(2000), 100000, 50)
}
func Benchmark5PercentCachePoolSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(5000), 100000, 50)
}

// Size 1000 Lambda 100
func BenchmarkNoCachePoolSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 1000, 100)
}
func BenchmarkP2PercentCachePoolSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(2), 1000, 100)
}
func BenchmarkP5PercentCachePoolSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(5), 1000, 100)
}
func Benchmark1PercentCachePoolSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(10), 1000, 100)
}
func Benchmark2PercentCachePoolSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(20), 1000, 100)
}
func Benchmark5PercentCachePoolSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(50), 1000, 100)
}

// Size 100000 Lambda 100
func BenchmarkNoCachePoolSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 100000, 100)
}
func BenchmarkP2PercentCachePoolSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(200), 100000, 100)
}
func BenchmarkP5PercentCachePoolSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(500), 100000, 100)
}
func Benchmark1PercentCachePoolSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(1000), 100000, 100)
}
func Benchmark2PercentCachePoolSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(2000), 100000, 100)
}
func Benchmark5PercentCachePoolSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCachePool(b, cdc.Cache(5000), 100000, 100)
}

func (pool *CachePool) logUnmarshalBinary(bz []byte, ptr interface{}) (hit int, err error) {
	cache := pool.acquireCache()
	hit, err = cache.logUnmarshalBinary(bz, ptr)
	pool.returnCache(cache)
	return
}

func (pool *CachePool) logUnmarshalJSON(bz []byte, ptr interface{}) (hit int, err error) {
	cache := pool.acquireCache()
	hit, err = cache.logUnmarshalJSON(bz, ptr)
	pool.returnCache(cache)
	return
}
