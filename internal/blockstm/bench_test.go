package blockstm

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/test-go/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func executeBlock(stores map[storetypes.StoreKey]int, storage MultiStore, worker int, block *MockBlock) error {
	incarnationCache := make([]atomic.Pointer[map[string]any], block.Size())
	for i := 0; i < block.Size(); i++ {
		m := make(map[string]any)
		incarnationCache[i].Store(&m)
	}
	return ExecuteBlock(context.Background(), block.Size(), stores, storage, worker, func(txn TxnIndex, store MultiStore) {
		cache := incarnationCache[txn].Swap(nil)
		block.ExecuteTx(txn, store, *cache)
		incarnationCache[txn].Store(cache)
	})
}

func BenchmarkBlockSTM(b *testing.B) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}
	for i := 0; i < 26; i++ {
		key := storetypes.NewKVStoreKey(strconv.FormatInt(int64(i), 10))
		stores[key] = i + 2
	}
	storage := NewMultiMemDB(stores)
	testCases := []struct {
		name  string
		block *MockBlock
	}{
		{"random-10000/100", testBlock(10000, 100)},
		{"no-conflict-10000", noConflictBlock(10000)},
		{"worst-case-10000", worstCaseBlock(10000)},
		{"iterate-10000/100", iterateBlock(10000, 100)},
	}
	for _, tc := range testCases {
		b.Run(tc.name+"-sequential", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runSequential(storage, tc.block)
			}
		})
		for _, worker := range []int{1, 5, 10, 15, 20} {
			b.Run(tc.name+"-worker-"+strconv.Itoa(worker), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					require.NoError(
						b, executeBlock(stores, storage, worker, tc.block),
					)
				}
			})
		}
	}
}

func runSequential(storage MultiStore, block *MockBlock) {
	for i, tx := range block.Txs {
		block.Results[i] = tx(storage, nil)
	}
}
