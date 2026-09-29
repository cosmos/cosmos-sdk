package blockstm

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

func TestExecuteBlock_CancelWakesSuspendedExecutors(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)

	// Mark key "k" as ESTIMATE for txn 0.
	estimates := make([]MultiLocations, 2)
	estimates[0] = MultiLocations{0: Locations{Key([]byte("k"))}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := ExecuteBlockWithEstimates(ctx, 2, stores, storage, 2, estimates,
		func(txn TxnIndex, store MultiStore) {
			if txn == 0 {
				time.Sleep(250 * time.Millisecond)
				return
			}
			// Txn 1 suspends on ESTIMATE.
			store.GetKVStore(StoreKeyAuth).Get([]byte("k"))
		},
	)
	require.True(t, errors.Is(err, context.Canceled))
}

// CancelAll must clear an Executing blocker's ESTIMATE, not just the
// Suspended waiter's — otherwise the woken waiter re-suspends on the same mark.
func TestCancelAllClearsBlockerEstimate(t *testing.T) {
	ctx := context.Background()
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}
	storage := NewMultiMemDB(stores)

	// Pre-estimate: tx1 has key "k" marked as ESTIMATE — it is the blocker.
	estimates := make([]MultiLocations, 3)
	estimates[1] = MultiLocations{0: Locations{Key([]byte("k"))}}

	scheduler := NewScheduler(3)
	mv := NewMVMemoryWithEstimates(3, stores, MultiStoreToStorage(storage, stores), scheduler, estimates)

	// tx1 is Executing (the blocker, NOT suspended).
	_, ok := scheduler.txnStatus[1].TrySetExecuting()
	require.True(t, ok)

	// tx2 is Suspended on tx1's ESTIMATE.
	_, ok = scheduler.txnStatus[2].TrySetExecuting()
	require.True(t, ok)
	cond := NewCondvar()
	scheduler.txnStatus[2].Suspend(cond)

	// Sanity: from tx2's perspective, key "k" currently reads as ESTIMATE.
	mvData := mv.data[0].(*MVData)
	_, _, isEstimate := mvData.Read(ctx, Key([]byte("k")), 2)
	require.True(t, isEstimate, "precondition: tx1's estimate should be visible to tx2")

	// Run the same callback the production code uses on ctx cancellation.
	scheduler.CancelAll(func(i TxnIndex) {
		mv.ClearEstimates(i)
	})

	// tx2 must be woken.
	cond.Lock()
	require.True(t, cond.notified)
	cond.Unlock()

	// Regression guard: blocker's ESTIMATE must be cleared too.
	_, _, isEstimate = mvData.Read(ctx, Key([]byte("k")), 2)
	require.False(t, isEstimate, "tx1's ESTIMATE must be cleared so tx2 doesn't re-suspend on resume")
}

func accountName(i int64) string {
	return fmt.Sprintf("account%05d", i)
}

func testBlock(size, accounts int) *MockBlock {
	txs := make([]Tx, size)
	g := rand.New(rand.NewSource(0))
	for i := 0; i < size; i++ {
		sender := g.Int63n(int64(accounts))
		receiver := g.Int63n(int64(accounts))
		txs[i] = BankTransferTx(i, accountName(sender), accountName(receiver), 1)
	}
	return NewMockBlock(txs)
}

func iterateBlock(size, accounts int) *MockBlock {
	txs := make([]Tx, size)
	g := rand.New(rand.NewSource(0))
	for i := 0; i < size; i++ {
		sender := g.Int63n(int64(accounts))
		receiver := g.Int63n(int64(accounts))
		txs[i] = IterateTx(i, accountName(sender), accountName(receiver), 1)
	}
	return NewMockBlock(txs)
}

func noConflictBlock(size int) *MockBlock {
	txs := make([]Tx, size)
	for i := 0; i < size; i++ {
		sender := accountName(int64(i))
		txs[i] = BankTransferTx(i, sender, sender, 1)
	}
	return NewMockBlock(txs)
}

func worstCaseBlock(size int) *MockBlock {
	txs := make([]Tx, size)
	for i := 0; i < size; i++ {
		// all transactions are from the same account
		sender := "account0"
		txs[i] = BankTransferTx(i, sender, sender, 1)
	}
	return NewMockBlock(txs)
}

func deterministicBlock() *MockBlock {
	return NewMockBlock([]Tx{
		NoopTx(0, "account0"),
		NoopTx(1, "account1"),
		NoopTx(2, "account1"),
		NoopTx(3, "account1"),
		NoopTx(4, "account3"),
		NoopTx(5, "account1"),
		NoopTx(6, "account4"),
		NoopTx(7, "account5"),
		NoopTx(8, "account6"),
	})
}

func TestSTM(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}
	testCases := []struct {
		name      string
		blk       *MockBlock
		executors int
	}{
		{
			name:      "testBlock(100,80),10",
			blk:       testBlock(100, 80),
			executors: 10,
		},
		{
			name:      "testBlock(100,3),10",
			blk:       testBlock(100, 3),
			executors: 10,
		},
		{
			name:      "determisticBlock(),5",
			blk:       deterministicBlock(),
			executors: 5,
		},
		{
			name:      "noConflictBlock(100),5",
			blk:       noConflictBlock(100),
			executors: 5,
		},
		{
			name:      "worstCaseBlock(100),5",
			blk:       worstCaseBlock(100),
			executors: 5,
		},
		{
			name:      "iterateBlock(100,80),10",
			blk:       iterateBlock(100, 80),
			executors: 10,
		},
		{
			name:      "iterateBlock(100,10),10",
			blk:       iterateBlock(100, 10),
			executors: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMultiMemDB(stores)
			require.NoError(t,
				ExecuteBlock(context.Background(), tc.blk.Size(), stores, storage, tc.executors, func(txn TxnIndex, store MultiStore) {
					tc.blk.ExecuteTx(txn, store, nil)
				}),
			)
			for _, err := range tc.blk.Results {
				require.NoError(t, err)
			}

			crossCheck := NewMultiMemDB(stores)
			runSequential(crossCheck, tc.blk)

			// check parallel execution matches sequential execution
			for store := range stores {
				require.True(t, StoreEqual(crossCheck.GetKVStore(store), storage.GetKVStore(store)))
			}

			// check total nonce increased the same amount as the number of transactions
			var total uint64
			store := storage.GetKVStore(StoreKeyAuth)
			it := store.Iterator(nil, nil)
			defer it.Close()

			for ; it.Valid(); it.Next() {
				if !bytes.HasPrefix(it.Key(), []byte("nonce")) {
					continue
				}
				total += binary.BigEndian.Uint64(it.Value())
			}
			require.Equal(t, uint64(tc.blk.Size()), total)
		})
	}
}

// TestSTMHighContentionStress runs high-contention blocks many times.
// With count=1, TestSTM passes because the bug is probabilistic.
// With 200 iterations, the scheduler non-determinism triggers reliably.
func TestSTMHighContentionStress(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}
	testCases := []struct {
		name      string
		blk       *MockBlock
		executors int
	}{
		{"worstCaseBlock(100),5", worstCaseBlock(100), 5},
		{"testBlock(100,3),10", testBlock(100, 3), 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 200; i++ {
				storage := NewMultiMemDB(stores)
				require.NoError(t,
					ExecuteBlock(context.Background(), tc.blk.Size(), stores, storage, tc.executors, func(txn TxnIndex, store MultiStore) {
						tc.blk.ExecuteTx(txn, store, nil)
					}),
				)
				for _, err := range tc.blk.Results {
					require.NoError(t, err)
				}

				crossCheck := NewMultiMemDB(stores)
				runSequential(crossCheck, tc.blk)

				for store := range stores {
					require.True(t, StoreEqual(crossCheck.GetKVStore(store), storage.GetKVStore(store)),
						"iteration %d: parallel != sequential for store %s", i, store.Name())
				}
			}
		})
	}
}

func TestValidateInputs(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}

	testCases := []struct {
		name      string
		blockSize int
		stores    map[storetypes.StoreKey]int
		estimates []MultiLocations
		errSubstr string // empty means expect no error
	}{
		{
			name:      "valid, no estimates",
			blockSize: 4,
			stores:    stores,
		},
		{
			name:      "valid with estimates",
			blockSize: 4,
			stores:    stores,
			estimates: []MultiLocations{0: {1: Locations{Key([]byte("k"))}}},
		},
		{
			name:      "zero block size is allowed",
			blockSize: 0,
			stores:    stores,
		},
		{
			name:      "negative block size",
			blockSize: -1,
			stores:    stores,
			errSubstr: "invalid block size",
		},
		{
			name:      "block size overflows uint32",
			blockSize: math.MaxUint32 + 1,
			stores:    stores,
			errSubstr: "overflows uint32",
		},
		{
			name:      "store index out of range",
			blockSize: 4,
			stores:    map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 2},
			errSubstr: "store index out of range",
		},
		{
			name:      "negative store index",
			blockSize: 4,
			stores:    map[storetypes.StoreKey]int{StoreKeyAuth: -1},
			errSubstr: "store index out of range",
		},
		{
			name:      "duplicate store index",
			blockSize: 4,
			stores:    map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 0},
			errSubstr: "duplicate store index",
		},
		{
			name:      "estimates longer than block",
			blockSize: 1,
			stores:    stores,
			estimates: make([]MultiLocations, 2),
			errSubstr: "exceeds block size",
		},
		{
			name:      "estimate store index out of range",
			blockSize: 4,
			stores:    stores,
			estimates: []MultiLocations{0: {5: Locations{Key([]byte("k"))}}},
			errSubstr: "references store index out of range",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMultiMemDB(stores)
			noop := func(TxnIndex, MultiStore) {}
			err := ExecuteBlockWithEstimates(context.Background(), tc.blockSize, tc.stores, storage, 1, tc.estimates, noop)
			if tc.errSubstr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errSubstr)
		})
	}
}

func StoreEqual(a, b storetypes.KVStore) bool {
	// compare with iterators
	iter1 := a.Iterator(nil, nil)
	iter2 := b.Iterator(nil, nil)
	defer iter1.Close()
	defer iter2.Close()

	for {
		if !iter1.Valid() && !iter2.Valid() {
			return true
		}
		if !iter1.Valid() || !iter2.Valid() {
			return false
		}
		if !bytes.Equal(iter1.Key(), iter2.Key()) || !bytes.Equal(iter1.Value(), iter2.Value()) {
			return false
		}
		iter1.Next()
		iter2.Next()
	}
}
