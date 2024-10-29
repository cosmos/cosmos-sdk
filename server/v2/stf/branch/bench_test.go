package branch

import (
	"testing"

	"cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
)

func Benchmark_CacheStack1_Set(b *testing.B) {
	bs := makeBranchStack(b, 1)
	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		err := bs.Set([]byte{0}, []byte{0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_CacheStack10_Set(b *testing.B) {
	bs := makeBranchStack(b, 10)
	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		err := bs.Set([]byte{0}, []byte{0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_CacheStack100_Set(b *testing.B) {
	bs := makeBranchStack(b, 100)
	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		err := bs.Set([]byte{0}, []byte{0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_CacheStack1_Get(b *testing.B) {
	bs := makeBranchStack(b, 1)
	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		_, err := bs.Get([]byte{0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_CacheStack10_Get(b *testing.B) {
	bs := makeBranchStack(b, 10)
	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		_, err := bs.Get([]byte{0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_CacheStack100_Get(b *testing.B) {
	bs := makeBranchStack(b, 100)
	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		_, err := bs.Get([]byte{0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_CacheStack1_Iterate(b *testing.B) {
	bs := makeBranchStack(b, 1)
	var keySink, valueSink any

	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		iter, err := bs.Iterator([]byte{0}, []byte{0})
		if err != nil {
			b.Fatal(err)
		}
		for iter.Valid() {
			iter.Next()
			keySink = iter.Key()
			valueSink = iter.Value()
		}
	}

	_ = keySink
	_ = valueSink
}

func Benchmark_CacheStack10_Iterate(b *testing.B) {
	bs := makeBranchStack(b, 10)
	var keySink, valueSink any

	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		iter, err := bs.Iterator([]byte{0}, []byte{0})
		if err != nil {
			b.Fatal(err)
		}
		for iter.Valid() {
			iter.Next()
			keySink = iter.Key()
			valueSink = iter.Value()
		}
	}

	_ = keySink
	_ = valueSink
}

func Benchmark_CacheStack100_Iterate(b *testing.B) {
	bs := makeBranchStack(b, 100)
	var keySink, valueSink any

	b.ResetTimer()
	b.ReportAllocs()
	// test sets
	for i := 0; i < b.N; i++ {
		iter, err := bs.Iterator([]byte{0}, []byte{0})
		if err != nil {
			b.Fatal(err)
		}
		for iter.Valid() {
			iter.Next()
			keySink = iter.Key()
			valueSink = iter.Value()
		}
	}

	_ = keySink
	_ = valueSink
}

func makeBranchStack(b *testing.B, stackSize int) Store[store.KVStore] {
	const elems = 10
	parent := coretesting.NewMemKV()
	branch := NewStore[store.KVStore](parent)
	for i := 1; i < stackSize; i++ {
		branch = NewStore[store.KVStore](branch)
		for j := 0; j < elems; j++ {
			key, value := []byte{byte(j)}, []byte{byte(j)}
			err := branch.Set(key, value)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	return branch
}
