package blockstm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func TestAccountCreationParallel_CacheWrapWithEstimatesDeterministic(t *testing.T) {
	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}

	const (
		numTxns    = 40
		workers    = 8
		iterations = 20
	)

	for iter := 0; iter < iterations; iter++ {
		txs := make([]Tx, numTxns)
		for i := 0; i < numTxns; i++ {
			txs[i] = accountCreationTx(i, accountCreationOpts{cacheWrap: true})
		}

		blk := NewMockBlock(txs)
		storage := NewMultiMemDB(stores)

		estimates := make([]MultiLocations, numTxns)
		for i := 0; i < numTxns; i++ {
			estimates[i] = MultiLocations{
				0: Locations{Key(accountCreationSeqKey)},
			}
		}

		err := ExecuteBlockWithEstimates(
			context.Background(), blk.Size(), stores, storage, workers, estimates,
			func(txn TxnIndex, store MultiStore) {
				blk.ExecuteTx(txn, store, nil)
			},
		)
		require.NoError(t, err, "iteration %d: ExecuteBlockWithEstimates failed", iter)

		for i, perr := range blk.Results {
			require.NoError(t, perr, "iteration %d: parallel tx %d failed: %v", iter, i, perr)
		}

		seqTxs := make([]Tx, numTxns)
		for i := 0; i < numTxns; i++ {
			seqTxs[i] = accountCreationTx(i, accountCreationOpts{cacheWrap: true})
		}
		seqBlk := NewMockBlock(seqTxs)
		crossCheck := NewMultiMemDB(stores)
		runSequential(crossCheck, seqBlk)

		for i, serr := range seqBlk.Results {
			require.NoError(t, serr, "iteration %d: sequential tx %d failed", iter, i)
		}

		require.True(
			t,
			StoreEqual(crossCheck.GetKVStore(StoreKeyAuth), storage.GetKVStore(StoreKeyAuth)),
			"iteration %d: parallel execution state diverges from sequential execution", iter,
		)

		require.Equal(
			t,
			uint64(numTxns),
			accountCreationFinalSeq(storage.GetKVStore(StoreKeyAuth)),
			"iteration %d: final sequence mismatch", iter,
		)
	}
}
