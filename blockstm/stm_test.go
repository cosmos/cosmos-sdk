package blockstm

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/test-go/testify/require"

	storetypes "cosmossdk.io/store/types"
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

func determisticBlock() *MockBlock {
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
			blk:       determisticBlock(),
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
				ExecuteBlock(context.Background(), tc.blk.Size(), stores, storage, tc.executors, tc.blk.ExecuteTx),
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
