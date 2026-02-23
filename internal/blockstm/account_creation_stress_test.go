package blockstm

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func runAccountCreationStressCase(
	t *testing.T,
	numTxns, workers, iterations int,
	opts accountCreationOpts,
	recoverPanics bool,
) {
	t.Helper()

	stores := map[storetypes.StoreKey]int{StoreKeyAuth: 0}

	for iter := 0; iter < iterations; iter++ {
		txs := make([]Tx, numTxns)
		for i := 0; i < numTxns; i++ {
			txs[i] = accountCreationTx(i, opts)
		}
		blk := NewMockBlock(txs)
		storage := NewMultiMemDB(stores)

		execTx := func(txn TxnIndex, store MultiStore) {
			if recoverPanics {
				defer func() {
					if r := recover(); r != nil {
						blk.Results[txn] = fmt.Errorf("panic recovered: %v", r)
					}
				}()
			}
			blk.ExecuteTx(txn, store, nil)
		}

		var err error
		estimates := make([]MultiLocations, numTxns)
		for i := 0; i < numTxns; i++ {
			estimates[i] = MultiLocations{
				0: Locations{Key(accountCreationSeqKey)},
			}
		}
		err = ExecuteBlockWithEstimates(
			context.Background(), blk.Size(), stores, storage, workers, estimates, execTx,
		)

		require.NoError(t, err, "iteration %d: ExecuteBlock failed", iter)

		seqTxs := make([]Tx, numTxns)
		for i := 0; i < numTxns; i++ {
			seqTxs[i] = accountCreationTx(i, opts)
		}
		seqBlk := NewMockBlock(seqTxs)
		crossCheck := NewMultiMemDB(stores)
		runSequential(crossCheck, seqBlk)

		for i, serr := range seqBlk.Results {
			require.NoError(t, serr, "iteration %d: sequential tx %d failed", iter, i)
		}
		for i, perr := range blk.Results {
			require.NoError(t, perr, "iteration %d: parallel tx %d failed", iter, i)
		}

		stateMatch := StoreEqual(crossCheck.GetKVStore(StoreKeyAuth), storage.GetKVStore(StoreKeyAuth))
		if !stateMatch {
			finalSeq := accountCreationFinalSeq(storage.GetKVStore(StoreKeyAuth))
			missing := accountCreationMissingUniqueNumbers(storage.GetKVStore(StoreKeyAuth), numTxns)
			t.Fatalf(
				"iteration %d: state diverged\n  final seq (want %d): %d\n  missing unique nums: %v",
				iter, numTxns, finalSeq, missing,
			)
		}

		require.Equal(
			t,
			uint64(numTxns),
			accountCreationFinalSeq(storage.GetKVStore(StoreKeyAuth)),
			"iteration %d: final sequence mismatch", iter,
		)
	}
}

func TestAccountCreationParallelRace(t *testing.T) {
	runAccountCreationStressCase(
		t,
		100,
		10,
		100,
		accountCreationOpts{},
		false,
	)
}

func TestAccountCreationParallel_WithEstimates(t *testing.T) {
	runAccountCreationStressCase(
		t,
		50,
		10,
		20,
		accountCreationOpts{},
		false,
	)
}

func TestAccountCreationParallel_DivergentPanicRecovery(t *testing.T) {
	runAccountCreationStressCase(
		t,
		50,
		10,
		20,
		accountCreationOpts{panicOnConflict: true},
		true,
	)
}

func TestAccountCreationParallel_CacheWrap(t *testing.T) {
	runAccountCreationStressCase(
		t,
		100,
		10,
		100,
		accountCreationOpts{cacheWrap: true},
		false,
	)
}

func TestAccountCreationParallel_CacheWrap_PanicPath(t *testing.T) {
	runAccountCreationStressCase(
		t,
		100,
		10,
		50,
		accountCreationOpts{cacheWrap: true, panicOnConflict: true},
		true,
	)
}

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
			require.NoError(t, perr, "iteration %d: parallel tx %d failed", iter, i)
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
