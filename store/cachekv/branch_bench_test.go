package cachekv_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/dbadapter"
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
				bs.Set([]byte{0}, []byte{0})
			}
		})
	}
}

// Gets the same key from the branch store.
func Benchmark_Get(b *testing.B) {
	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sink = bs.Get([]byte{0})
			}
		})
	}
	if sink == nil {
		b.Fatal("prevent compiler optimization")
	}
	sink = nil
}

// Gets always different keys.
func Benchmark_GetSparse(b *testing.B) {
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
				sink = bs.Get(key)
			}
		})
	}
	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = nil
}

var keySink, valueSink any

func Benchmark_Iterate(b *testing.B) {
	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				iter := bs.Iterator(nil, nil)
				for iter.Valid() {
					keySink = iter.Key()
					valueSink = iter.Value()
					iter.Next()
				}
				_ = iter.Close()
			}
		})
	}

	if keySink == nil || valueSink == nil {
		b.Fatal("Benchmark did not run")
	}
	keySink = nil
	valueSink = nil
}

// makeBranchStack creates a branch stack of the given size and initializes it with unique key-value pairs.
func makeBranchStack(_ *testing.B, stackSize int) *cachekv.Store {
	parent := dbadapter.Store{DB: coretesting.NewMemDB()}
	branch := cachekv.NewStore(parent)
	for i := 1; i < stackSize; i++ {
		branch = cachekv.NewStore(branch)
		for j := 0; j < elemsInStack; j++ {
			// create unique keys by including the branch index.
			key := append(numToBytes(i), numToBytes(j)...)
			value := []byte{byte(j)}
			branch.Set(key, value)
		}
	}
	return branch
}

func numToBytes[T ~int](n T) []byte {
	return binary.BigEndian.AppendUint64(nil, uint64(n))
}
