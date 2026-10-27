package blockstm

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

// counterOnlyBlock is the minimal maximum-contention block: every transaction
// does a read-modify-write on a single hot key. A correct serial (and any
// serializable parallel) execution of N such transactions must leave the
// counter at exactly N.
func counterOnlyBlock(size int) *MockBlock {
	txs := make([]Tx, size)
	for i := range size {
		txs[i] = func(ms MultiStore, _ Cache) error {
			acc := ms.GetKVStore(StoreKeyAuth)
			ctr := []byte("counter")
			var n uint64
			if v := acc.Get(ctr); v != nil {
				n = binary.BigEndian.Uint64(v)
			}
			n++
			var bz [8]byte
			binary.BigEndian.PutUint64(bz[:], n)
			acc.Set(ctr, bz[:])
			return nil
		}
	}
	return NewMockBlock(txs)
}

// TestSchedulerNoLostUpdates is a regression test for a scheduler bug where the
// validation task path incremented numActiveTasks AFTER advancing validationIdx,
// opening a window in which CheckDone could latch doneMarker=true while a
// validation (that would abort a stale-read txn and reschedule work) was still
// pending. The dropped re-executions committed stale reads => lost updates.
//
// The bug is a rare interleaving; it reproduces most reliably under `-race`,
// which perturbs goroutine scheduling. Before the fix, the counter dropped
// below blockSize within a few hundred iterations.
func TestSchedulerNoLostUpdates(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}
	const blockSize = 80
	for iter := range 1500 {
		blk := counterOnlyBlock(blockSize)
		storage := NewMultiMemDB(stores)
		require.NoError(t, ExecuteBlock(context.Background(), blk.Size(), stores, storage, 8,
			func(txn TxnIndex, store MultiStore) { blk.ExecuteTx(txn, store, nil) }))

		got := binary.BigEndian.Uint64(storage.GetKVStore(StoreKeyAuth).Get([]byte("counter")))
		require.Equalf(t, uint64(blockSize), got,
			"iter %d: counter=%d, want %d (lost %d committed updates)",
			iter, got, blockSize, blockSize-int(got))
	}
}
