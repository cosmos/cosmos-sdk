package blockstm

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/test-go/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func BenchmarkBlockSTM(b *testing.B) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}
	for i := 0; i < 26; i++ {
		key := storetypes.NewKVStoreKey(strconv.FormatInt(int64(i), 10))
		stores[key] = i + 2
	}
	abasamevalueKeys := abaSameValueKeys(10000)
	hasKeys := hasKeys(10000)
	abaBigValue := make([]byte, 64<<10) // 64KiB
	iterateAccounts := 100
	testCases := []struct {
		name  string
		block *MockBlock
		setup func(MultiStore)
	}{
		{"random-10000/100", testBlock(10000, 100), nil},
		{"has-hit-10000/100", hasBlock(10000, 100, hasKeys), func(storage MultiStore) { prepopulateHasKeys(storage, hasKeys) }},
		{"has-miss-10000/100", hasBlock(10000, 100, hasKeys), nil},
		{"no-conflict-10000", noConflictBlock(10000), nil},
		{"worst-case-10000", worstCaseBlock(10000), nil},
		{"iterate-10000/100", iterateBlock(10000, iterateAccounts), nil},
		{
			"iterate-10000/100-prepop",
			iterateBlock(10000, iterateAccounts),
			func(storage MultiStore) { prepopulateIterateAccounts(storage, iterateAccounts) },
		},
		{"iterate-newkeys-2000", iterateNewKeysBlock(2000), nil},
		{
			"aba-samevalue-10000",
			abaSameValueBlock(abasamevalueKeys),
			func(storage MultiStore) { prepopulateABASameValue(storage, abasamevalueKeys) },
		},
		{
			"aba-samevalue-bigvalue-10000",
			abaSameValueBlockWithValue(abasamevalueKeys, abaBigValue),
			func(storage MultiStore) { prepopulateABASameValueWithValue(storage, abasamevalueKeys, abaBigValue) },
		},
	}
	for _, tc := range testCases {
		b.Run(tc.name+"-sequential", func(b *testing.B) {
			storage := NewMultiMemDB(stores)
			if tc.setup != nil {
				tc.setup(storage)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runSequential(storage, tc.block)
			}
			b.ReportMetric(1, "exec/txn")
			b.ReportMetric(0, "val/txn")
		})
		for _, worker := range []int{1, 5, 10, 15, 20} {
			b.Run(tc.name+"-worker-"+strconv.Itoa(worker), func(b *testing.B) {
				storage := NewMultiMemDB(stores)
				if tc.setup != nil {
					tc.setup(storage)
				}

				incarnationCache := make([]atomic.Pointer[map[string]any], tc.block.Size())
				for i := 0; i < tc.block.Size(); i++ {
					m := make(map[string]any)
					incarnationCache[i].Store(&m)
				}

				b.ResetTimer()
				var executedTotal, validatedTotal uint64
				for i := 0; i < b.N; i++ {
					executed, validated, err := executeBlockWithEstimatesImpl(
						context.Background(),
						tc.block.Size(),
						stores,
						storage,
						worker,
						nil,
						func(txn TxnIndex, store MultiStore) {
							cache := incarnationCache[txn].Swap(nil)
							tc.block.ExecuteTx(txn, store, *cache)
							incarnationCache[txn].Store(cache)
						},
						false,
					)
					require.NoError(b, err)
					executedTotal += executed
					validatedTotal += validated
				}
				denom := float64(b.N * tc.block.Size())
				b.ReportMetric(float64(executedTotal)/denom, "exec/txn")
				b.ReportMetric(float64(validatedTotal)/denom, "val/txn")
			})
		}
	}
}

func prepopulateIterateAccounts(storage MultiStore, accounts int) {
	auth := storage.GetKVStore(StoreKeyAuth)
	bank := storage.GetKVStore(StoreKeyBank)
	zero := make([]byte, 8)
	for i := 0; i < accounts; i++ {
		acc := accountName(int64(i))
		auth.Set([]byte("nonce"+acc), zero)
		bank.Set([]byte("balance"+acc), zero)
	}
}

// iterateNewKeysBlock stresses unordered index iteration costs by inserting a new key in each
// transaction and immediately iterating the store, forcing frequent key snapshot rebuilds.
func iterateNewKeysBlock(size int) *MockBlock {
	txs := make([]Tx, size)
	for i := 0; i < size; i++ {
		idx := i
		txs[i] = func(store MultiStore, _ Cache) error {
			kv := store.GetKVStore(StoreKeyAuth)
			kv.Set([]byte(fmt.Sprintf("iter-newkeys/%08d", idx)), []byte{1})

			it := kv.Iterator(nil, nil)
			defer it.Close()
			for j := 0; it.Valid() && j < 10; j++ {
				it.Next()
			}
			return nil
		}
	}
	return NewMockBlock(txs)
}

func runSequential(storage MultiStore, block *MockBlock) {
	for i, tx := range block.Txs {
		block.Results[i] = tx(storage, nil)
	}
}

func abaSameValueKeys(n int) [][]byte {
	keys := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = []byte(fmt.Sprintf("aba-samevalue/%08d", i))
	}
	return keys
}

// prepopulateABASameValue seeds storage with values matching the transaction writes.
// This creates a scenario where value-based validation prevents unnecessary re-execution.
func prepopulateABASameValue(storage MultiStore, keys [][]byte) {
	kv := storage.GetKVStore(StoreKeyAuth)
	value := []byte{1}
	for _, key := range keys {
		kv.Set(key, value)
	}
}

func prepopulateABASameValueWithValue(storage MultiStore, keys [][]byte, value []byte) {
	kv := storage.GetKVStore(StoreKeyAuth)
	for _, key := range keys {
		kv.Set(key, value)
	}
}

func hasKeys(n int) [][]byte {
	keys := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = []byte(fmt.Sprintf("has/%08d", i))
	}
	return keys
}

func prepopulateHasKeys(storage MultiStore, keys [][]byte) {
	kv := storage.GetKVStore(StoreKeyAuth)
	value := []byte{1}
	for _, key := range keys {
		kv.Set(key, value)
	}
}

func hasBlock(size, readsPerTx int, keys [][]byte) *MockBlock {
	txs := make([]Tx, size)
	for i := 0; i < size; i++ {
		idx := i
		txs[i] = func(store MultiStore, _ Cache) error {
			kv := store.GetKVStore(StoreKeyAuth)
			for j := 0; j < readsPerTx; j++ {
				_ = kv.Has(keys[(idx+j)%len(keys)])
			}
			return nil
		}
	}
	return NewMockBlock(txs)
}

func abaSameValueBlock(keys [][]byte) *MockBlock {
	value := []byte{1}
	return abaSameValueBlockWithValue(keys, value)
}

func abaSameValueBlockWithValue(keys [][]byte, value []byte) *MockBlock {
	txs := make([]Tx, len(keys))
	for i := range keys {
		idx := i
		txs[i] = func(store MultiStore, _ Cache) error {
			kv := store.GetKVStore(StoreKeyAuth)
			// Read dependency on previous key.
			if idx > 0 {
				_ = kv.Get(keys[idx-1])
			}
			// Write same value as pre-state.
			kv.Set(keys[idx], value)
			return nil
		}
	}
	return NewMockBlock(txs)
}
