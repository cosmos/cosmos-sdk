package blockstm

import (
	"context"
	"sync"
	"sync/atomic"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/collections"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.TxRunner = (*STMRunner)(nil)

func NewSTMRunner(
	txDecoder sdk.TxDecoder,
	stores []storetypes.StoreKey,
	workers int, estimate bool,
	coinDenom func(storetypes.MultiStore) string,
) *STMRunner {
	return &STMRunner{
		txDecoder: txDecoder,
		stores:    stores,
		workers:   workers,
		estimate:  estimate,
		coinDenom: coinDenom,
	}
}

// STMRunner simple implementation of block-stm
type STMRunner struct {
	txDecoder sdk.TxDecoder
	stores    []storetypes.StoreKey
	workers   int
	estimate  bool
	coinDenom func(storetypes.MultiStore) string

	// lastDebug holds the debug data from the most recent block execution.
	lastDebug atomic.Pointer[BlockExecutionDebug]
}

func (e *STMRunner) Run(ctx context.Context, ms storetypes.MultiStore, txs [][]byte, deliverTx sdk.DeliverTxFunc) ([]*abci.ExecTxResult, error) {
	var authStore, bankStore int
	index := make(map[storetypes.StoreKey]int, len(e.stores))
	for i, k := range e.stores {
		switch k.Name() {
		case "acc":
			authStore = i
		case "bank":
			bankStore = i
		}
		index[k] = i
	}

	blockSize := len(txs)
	if blockSize == 0 {
		return nil, nil
	}
	results := make([]*abci.ExecTxResult, blockSize)
	incarnationCache := make([]atomic.Pointer[map[string]any], blockSize)
	for i := 0; i < blockSize; i++ {
		m := make(map[string]any)
		incarnationCache[i].Store(&m)
	}

	var (
		estimates []MultiLocations
		memTxs    []sdk.Tx
	)

	if e.estimate {
		memTxs, estimates = preEstimates(txs, e.workers, authStore, bankStore, e.coinDenom(ms), e.txDecoder)
	}

	debug, err := ExecuteBlockWithEstimates(
		ctx,
		blockSize,
		index,
		stmMultiStoreWrapper{ms},
		e.workers,
		estimates,
		func(txn TxnIndex, ms MultiStore) {
			var cache map[string]any

			// only one of the concurrent incarnations gets the cache if there are any, otherwise execute without
			// cache, concurrent incarnations should be rare.
			v := incarnationCache[txn].Swap(nil)
			if v != nil {
				cache = *v
			}

			var memTx sdk.Tx
			if memTxs != nil {
				memTx = memTxs[txn]
			}
			results[txn] = deliverTx(txs[txn], memTx, msWrapper{ms}, int(txn), cache)

			if v != nil {
				incarnationCache[txn].Store(v)
			}
		},
	)
	if debug != nil {
		e.lastDebug.Store(debug)
	}
	if err != nil {
		return nil, err
	}

	return results, nil
}

// LastDebug returns the debug data from the most recent block execution, or nil if none.
func (e *STMRunner) LastDebug() *BlockExecutionDebug {
	return e.lastDebug.Load()
}

// SaveLastDebug persists the most recent block execution debug data to a file.
// Returns nil if there is no debug data available.
func (e *STMRunner) SaveLastDebug(path string) error {
	debug := e.lastDebug.Load()
	if debug == nil {
		return nil
	}
	return debug.SaveToFile(path)
}

// preEstimates returns a static estimation of the written keys for each transaction.
// NOTE: make sure it sync with the latest sdk logic when sdk upgrade.
func preEstimates(txs [][]byte, workers, authStore, bankStore int, coinDenom string, txDecoder sdk.TxDecoder) ([]sdk.Tx, []MultiLocations) {
	memTxs := make([]sdk.Tx, len(txs))
	estimates := make([]MultiLocations, len(txs))

	job := func(start, end int) {
		for i := start; i < end; i++ {
			rawTx := txs[i]
			tx, err := txDecoder(rawTx)
			if err != nil {
				continue
			}
			memTxs[i] = tx

			feeTx, ok := tx.(sdk.FeeTx)
			if !ok {
				continue
			}
			feePayer := sdk.AccAddress(feeTx.FeePayer())

			// account key
			accKey, err := collections.EncodeKeyWithPrefix(
				collections.NewPrefix(1),
				sdk.AccAddressKey,
				feePayer,
			)
			if err != nil {
				continue
			}

			// balance key
			balanceKey, err := collections.EncodeKeyWithPrefix(
				collections.NewPrefix(2),
				collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
				collections.Join(feePayer, coinDenom),
			)
			if err != nil {
				continue
			}

			estimates[i] = MultiLocations{
				authStore: {accKey},
				bankStore: {balanceKey},
			}
		}
	}

	blockSize := len(txs)
	chunk := (blockSize + workers - 1) / workers
	var wg sync.WaitGroup
	for i := 0; i < blockSize; i += chunk {
		start := i
		end := min(i+chunk, blockSize)
		wg.Add(1)
		go func() {
			defer wg.Done()
			job(start, end)
		}()
	}
	wg.Wait()

	return memTxs, estimates
}
