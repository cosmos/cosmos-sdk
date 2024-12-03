package branch

import (
	"encoding/binary"
	"fmt"
	"testing"

	"cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
)

var (
	stackSizes   = []int{1, 10, 100}
	elemsInStack = 10
)

func Benchmark_CacheStack_Set(b *testing.B) {
	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				err := bs.Set([]byte{0}, []byte{0})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

var sink any

func Benchmark_Get(b *testing.B) {
	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sink, _ = bs.Get([]byte{0})
			}
		})
	}
	if sink == nil {
		b.Fatal("benchmark did not run")
	}
	sink = nil
}

func Benchmark_GetSparse(b *testing.B) {
	var sink any
	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			keys := func() [][]byte {
				var keys [][]byte
				for i := 0; i < b.N; i++ {
					keys = append(keys, numToBytes(i))
				}
				return keys
			}()
			b.ResetTimer()
			b.ReportAllocs()
			for _, key := range keys {
				sink, _ = bs.Get(key)
			}
		})
	}
	if sink == nil {
		b.Fatal("benchmark did not run")
	}
	sink = nil
}

var (
	keySink   any
	valueSink any
)

func Benchmark_Iterate(b *testing.B) {
	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				iter, _ := bs.Iterator(nil, nil)
				for iter.Valid() {
					keySink = iter.Key()
					valueSink = iter.Value()
					iter.Next()
				}
				_ = iter.Close()
			}
		})
	}
	if valueSink == nil || keySink == nil {
		b.Fatal("benchmark did not run")
	}
	valueSink = nil
	keySink = nil
}

// makeBranchStack creates a branch stack of the given size and initializes it with unique key-value pairs.
func makeBranchStack(b *testing.B, stackSize int) Store[store.KVStore] {
	b.Helper()
	parent := coretesting.NewMemKV()
	branch := NewStore[store.KVStore](parent)
	for i := 1; i < stackSize; i++ {
		branch = NewStore[store.KVStore](branch)
		for j := 0; j < elemsInStack; j++ {
			// create unique keys by including the branch index.
			key := append(numToBytes(i), numToBytes(j)...)
			value := []byte{byte(j)}
			err := branch.Set(key, value)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	return branch
}

func numToBytes[T ~int](n T) []byte {
	return binary.BigEndian.AppendUint64(nil, uint64(n))
}
