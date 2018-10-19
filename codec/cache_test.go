package codec

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

type s1 struct {
	i int
}

type s2 struct {
	i interface{}
}

type ss struct {
	b bool
	i int64
	u uint64
}

type complex struct {
	s1 s1
	s2 s2
	ss ss
}

func newComplex() complex {
	return complex{
		s1: s1{rand.Int()},
		s2: s2{s1{rand.Int()}},
		ss: ss{
			b: false,
			i: rand.Int63(),
			u: rand.Uint64(),
		},
	}
}

type tc struct {
	value interface{}
	ptr   func() interface{}
}

func randgen(size int) (res []tc) {
	res = make([]tc, size)
	for i := range res {
		switch rand.Int() % 5 {
		case 0:
			res[i] = tc{rand.Int63(), func() interface{} { return new(int64) }}
		case 1:
			bz := make([]byte, 32)
			rand.Read(bz)
			res[i] = tc{bz, func() interface{} { return new([]byte) }}
		case 2:
			res[i] = tc{&s1{rand.Int()}, func() interface{} { return new(*s1) }}
		case 3:
			res[i] = tc{s2{s1{rand.Int()}}, func() interface{} { return new(s2) }}
		case 4:
			res[i] = tc{newComplex(), func() interface{} { return new(complex) }}
		}
	}
	return
}

func registerCodec(cdc *Amino) {
	cdc.RegisterConcrete(s1{}, "test/s1", nil)
	cdc.RegisterConcrete(s2{}, "test/s2", nil)
}

type marshaller func(interface{}) ([]byte, error)
type unmarshaller func([]byte, interface{}) error
type mustMarshaller func(interface{}) []byte
type mustUnmarshaller func([]byte, interface{})

func testEqual(t *testing.T, tc tc, marshal1 marshaller, unmarshal1 unmarshaller, marshal2 marshaller, unmarshal2 unmarshaller) {
	bz1, err := marshal1(tc.value)
	require.NoError(t, err)
	ptr1 := tc.ptr()
	err = unmarshal1(bz1, ptr1)
	require.NoError(t, err)

	bz2, err := marshal2(tc.value)
	require.NoError(t, err)
	ptr2 := tc.ptr()
	err = unmarshal2(bz2, ptr2)
	require.NoError(t, err)

	require.Equal(t, bz1, bz2)
	require.Equal(t, ptr1, ptr2)
}

func testEqualMust(t *testing.T, tc tc, marshal1 mustMarshaller, unmarshal1 mustUnmarshaller, marshal2 mustMarshaller, unmarshal2 mustUnmarshaller) {
	bz1 := marshal1(tc.value)
	ptr1 := tc.ptr()
	unmarshal1(bz1, ptr1)

	bz2 := marshal2(tc.value)
	ptr2 := tc.ptr()
	unmarshal2(bz2, ptr2)

	require.Equal(t, bz1, bz2)
	require.Equal(t, ptr1, ptr2)
}

/*
func TestIdentity(t *testing.T) {
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
*/
func benchmarkCodec(b *testing.B, cdc Codec, datanum int, lambda int) {
	tcs := randgen(datanum)
	b.ResetTimer()

	switch cdc := cdc.(type) {
	case *Amino:
		for i := 0; i < b.N; i++ {
			index := int(rand.ExpFloat64()*float64(datanum)/float64(lambda)) % datanum
			tc := tcs[index]
			ptr := tc.ptr()

			bz, _ := cdc.MarshalJSON(tc.value)
			cdc.UnmarshalJSON(bz, ptr)

			bz, _ = cdc.MarshalBinary(tc.value)
			cdc.UnmarshalBinary(bz, ptr)
		}
	case *Cache:
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

}

// Size 1000 Lambda 10
func BenchmarkNoCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc, 1000, 10)
}
func BenchmarkP2PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(2), 1000, 10)
}
func BenchmarkP5PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(5), 1000, 10)
}
func Benchmark1PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(10), 1000, 10)
}
func Benchmark2PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(20), 1000, 10)
}
func Benchmark5PercentCacheSize1000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(50), 1000, 10)
}

// Size 100000 Lambda 10
func BenchmarkNoCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc, 100000, 10)
}
func BenchmarkP2PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(200), 100000, 10)
}
func BenchmarkP5PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(500), 100000, 10)
}
func Benchmark1PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(1000), 100000, 10)
}
func Benchmark2PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(2000), 100000, 10)
}
func Benchmark5PercentCacheSize10000Lambda10(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(5000), 100000, 10)
}

// Size 1000 Lambda 50
func BenchmarkNoCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc, 1000, 50)
}
func BenchmarkP2PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(2), 1000, 50)
}
func BenchmarkP5PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(5), 1000, 50)
}
func Benchmark1PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(10), 1000, 50)
}
func Benchmark2PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(20), 1000, 50)
}
func Benchmark5PercentCacheSize1000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(50), 1000, 50)
}

// Size 100000 Lambda 50
func BenchmarkNoCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc, 100000, 50)
}
func BenchmarkP2PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(200), 100000, 50)
}
func BenchmarkP5PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(500), 100000, 50)
}
func Benchmark1PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(1000), 100000, 50)
}
func Benchmark2PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(2000), 100000, 50)
}
func Benchmark5PercentCacheSize10000Lambda50(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(5000), 100000, 50)
}

// Size 1000 Lambda 100
func BenchmarkNoCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc, 1000, 100)
}
func BenchmarkP2PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(2), 1000, 100)
}
func BenchmarkP5PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(5), 1000, 100)
}
func Benchmark1PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(10), 1000, 100)
}
func Benchmark2PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(20), 1000, 100)
}
func Benchmark5PercentCacheSize1000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(50), 1000, 100)
}

// Size 100000 Lambda 100
func BenchmarkNoCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc, 100000, 100)
}
func BenchmarkP2PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(200), 100000, 100)
}
func BenchmarkP5PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(500), 100000, 100)
}
func Benchmark1PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(1000), 100000, 100)
}
func Benchmark2PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(2000), 100000, 100)
}
func Benchmark5PercentCacheSize10000Lambda100(b *testing.B) {
	cdc := New()
	registerCodec(cdc)
	cdc.Seal()
	benchmarkCodec(b, cdc.Cache(5000), 100000, 100)
}

func (c *Cache) logUnmarshalBinary(bz []byte, ptr interface{}) (hit int, err error) {
	lru := c.bin
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

func (c *Cache) logUnmarshalJSON(bz []byte, ptr interface{}) (hit int, err error) {
	lru := c.json
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
