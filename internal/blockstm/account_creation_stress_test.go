package blockstm

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func requireBlockSTMStress(t *testing.T) {
	t.Helper()
	if os.Getenv("BLOCKSTM_STRESS") != "1" {
		t.Skip("set BLOCKSTM_STRESS=1 to run stress account-creation repro tests")
	}
}

func blockSTMStressIterations(defaultIterations int) int {
	if raw := os.Getenv("BLOCKSTM_STRESS_ITERS"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			return v
		}
	}
	return defaultIterations
}

func runAccountCreationStressCase(
	t *testing.T,
	numTxns, workers, iterations int,
	opts accountCreationOpts,
	useEstimates bool,
	recoverPanics bool,
) {
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
		if useEstimates {
			estimates := make([]MultiLocations, numTxns)
			for i := 0; i < numTxns; i++ {
				estimates[i] = MultiLocations{
					0: Locations{Key(accountCreationSeqKey)},
				}
			}
			err = ExecuteBlockWithEstimates(
				context.Background(), blk.Size(), stores, storage, workers, estimates, execTx,
			)
		} else {
			err = ExecuteBlock(
				context.Background(), blk.Size(), stores, storage, workers, execTx,
			)
		}

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
			require.NoError(t, perr, "iteration %d: parallel tx %d failed: %v", iter, i, perr)
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
	requireBlockSTMStress(t)
	runAccountCreationStressCase(
		t,
		100,
		10,
		blockSTMStressIterations(100),
		accountCreationOpts{},
		false,
		false,
	)
}

func TestAccountCreationParallel_WithEstimates(t *testing.T) {
	requireBlockSTMStress(t)
	runAccountCreationStressCase(
		t,
		50,
		10,
		blockSTMStressIterations(20),
		accountCreationOpts{},
		true,
		false,
	)
}

func TestAccountCreationParallel_DivergentPanicRecovery(t *testing.T) {
	requireBlockSTMStress(t)
	runAccountCreationStressCase(
		t,
		50,
		10,
		blockSTMStressIterations(20),
		accountCreationOpts{panicOnConflict: true},
		false,
		true,
	)
}

func TestAccountCreationParallel_CacheWrap(t *testing.T) {
	requireBlockSTMStress(t)
	runAccountCreationStressCase(
		t,
		100,
		10,
		blockSTMStressIterations(100),
		accountCreationOpts{cacheWrap: true},
		false,
		false,
	)
}

func TestAccountCreationParallel_CacheWrap_PanicPath(t *testing.T) {
	requireBlockSTMStress(t)
	runAccountCreationStressCase(
		t,
		100,
		10,
		blockSTMStressIterations(50),
		accountCreationOpts{cacheWrap: true, panicOnConflict: true},
		false,
		true,
	)
}
