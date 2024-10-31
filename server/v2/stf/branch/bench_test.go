package branch

import (
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

func Benchmark_Get(b *testing.B) {
	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := bs.Get([]byte{0})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func Benchmark_Iterate(b *testing.B) {
	var keySink, valueSink any

	for _, stackSize := range stackSizes {
		b.Run(fmt.Sprintf("StackSize%d", stackSize), func(b *testing.B) {
			bs := makeBranchStack(b, stackSize)
			b.ResetTimer()
			b.ReportAllocs()
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
				err = iter.Close()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	_ = keySink
	_ = valueSink
}

// makeBranchStack creates a branch stack of the given size and initializes it with unique key-value pairs.
func makeBranchStack(b *testing.B, stackSize int) Store[store.KVStore] {
	parent := coretesting.NewMemKV()
	branch := NewStore[store.KVStore](parent)
	for i := 1; i < stackSize; i++ {
		branch = NewStore[store.KVStore](branch)
		for j := 0; j < elemsInStack; j++ {
			// create unique keys by including the branch index.
			key := []byte{byte(i), byte(j)}
			value := []byte{byte(j)}
			err := branch.Set(key, value)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	return branch
}
