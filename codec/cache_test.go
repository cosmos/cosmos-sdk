package codec

import (
	"math/rand"
	"testing"
)

func TestCacheIdentity(t *testing.T) {
	size := 100

	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	cache := newCache(cdc, size)

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

func benchmarkCache(b *testing.B, cdc *cache, datanum, lambda int) {
	tcs := randgen(datanum)
	b.ResetTimer()

	var chit int
	exec := executor()
	for i := 0; i < b.N; i++ {

		index := int(rand.ExpFloat64()*float64(datanum)/float64(lambda)) % datanum
		tc := tcs[index]
		ptr := tc.ptr()

		exec(func() {
			bz, _ := cdc.MarshalJSON(tc.value)
			hit, _ := cdc.logUnmarshalJSON(bz, ptr)
			chit += hit
		})

		exec(func() {
			bz, _ := cdc.MarshalBinary(tc.value)
			hit, _ := cdc.logUnmarshalBinary(bz, ptr)
			chit += hit
		})
	}
	//	b.Logf("Cache hit %.2f%%", float64(chit)/2/float64(b.N)*100)
}

/*
// Size 1000 Lambda 10
func BenchmarkNoCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 1000, 10)
}
func BenchmarkP2PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 2), 1000, 10)
}
func BenchmarkP5PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 5), 1000, 10)
}
func Benchmark1PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 10), 1000, 10)
}
func Benchmark2PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 20), 1000, 10)
}
func Benchmark5PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 50), 1000, 10)
}

// Size 100000 Lambda 10
func BenchmarkNoCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 100000, 10)
}
func BenchmarkP2PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 200), 100000, 10)
}
func BenchmarkP5PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 500), 100000, 10)
}
func Benchmark1PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 1000), 100000, 10)
}
func Benchmark2PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 2000), 100000, 10)
}
func Benchmark5PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 5000), 100000, 10)
}

// Size 1000 Lambda 50
func BenchmarkNoCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 1000, 50)
}
func BenchmarkP2PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 2), 1000, 50)
}
func BenchmarkP5PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 5), 1000, 50)
}
func Benchmark1PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 10), 1000, 50)
}
func Benchmark2PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 20), 1000, 50)
}
func Benchmark5PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 50), 1000, 50)
}

// Size 100000 Lambda 50
func BenchmarkNoCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 100000, 50)
}
func BenchmarkP2PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 200), 100000, 50)
}
func BenchmarkP5PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 500), 100000, 50)
}
func Benchmark1PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 1000), 100000, 50)
}
func Benchmark2PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 2000), 100000, 50)
}
func Benchmark5PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 5000), 100000, 50)
}

// Size 1000 Lambda 100
func BenchmarkNoCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 1000, 100)
}
func BenchmarkP2PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 2), 1000, 100)
}
func BenchmarkP5PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 5), 1000, 100)
}
func Benchmark1PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 10), 1000, 100)
}
func Benchmark2PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 20), 1000, 100)
}
func Benchmark5PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 50), 1000, 100)
}

// Size 100000 Lambda 100
func BenchmarkNoCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkAmino(b, cdc, 100000, 100)
}
func BenchmarkP2PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 200), 100000, 100)
}
func BenchmarkP5PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 500), 100000, 100)
}
func Benchmark1PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 1000), 100000, 100)
}
func Benchmark2PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 2000), 100000, 100)
}
func Benchmark5PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCache(b, newCache(cdc, 5000), 100000, 100)
}
*/
func (c *cache) logUnmarshalBinary(bz []byte, ptr interface{}) (hit int, err error) {
	lru := c.bin
	lru.lock()
	defer lru.unlock()
	if lru.read(bz, ptr) {
		return 1, nil
	}
	err = c.cdc.UnmarshalBinary(bz, ptr)
	if err != nil {
		return
	}
	lru.write(bz, ptr)
	return
}

func (c *cache) logUnmarshalJSON(bz []byte, ptr interface{}) (hit int, err error) {
	lru := c.json
	lru.lock()
	defer lru.unlock()
	if lru.read(bz, ptr) {
		return 1, nil
	}
	err = c.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		return
	}
	lru.write(bz, ptr)
	return
}
