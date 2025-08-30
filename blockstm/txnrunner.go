package blockstm

import (
	"context"
	"sync"
	"sync/atomic"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewDefaultRunner(ctx context.Context, txDecoder sdk.TxDecoder, deliverTx func([]byte) *abci.ExecTxResult) *DefaultRunner {
	return &DefaultRunner{
		ctx:       ctx,
		txDecoder: txDecoder,
		deliverTx: deliverTx,
	}
}

// DefaultRunner default executor without parallelism
type DefaultRunner struct {
	ctx       context.Context
	txDecoder sdk.TxDecoder
	deliverTx func(tx []byte) *abci.ExecTxResult
}

func (d DefaultRunner) Run(txs [][]byte) ([]*abci.ExecTxResult, error) {
	// Fallback to the default execution logic
	txResults := make([]*abci.ExecTxResult, 0, len(txs))
	for _, rawTx := range txs {
		var response *abci.ExecTxResult

		if _, err := d.txDecoder(rawTx); err == nil {
			response = d.deliverTx(rawTx)
		} else {
			// In the case where a transaction included in a block proposal is malformed,
			// we still want to return a default response to comet. This is because comet
			// expects a response for each transaction included in a block proposal.
			response = sdkerrors.ResponseExecTxResultWithEvents(
				sdkerrors.ErrTxDecode,
				0,
				0,
				nil,
				false,
			)
		}

		// check after every tx if we should abort
		select {
		case <-d.ctx.Done():
			return nil, d.ctx.Err()
		default:
			// continue
		}

		txResults = append(txResults, response)
	}
	return txResults, nil
}

// STMRunner simple implementation of block-stm
type STMRunner struct {
	ctx       context.Context
	txDecoder sdk.TxDecoder
	stores    []storetypes.StoreKey
	ms        MultiStore
	workers   int
	estimate  bool
	coinDenom string
	deliverTx func(int, sdk.Tx, MultiStore, map[string]any) *abci.ExecTxResult
}

func (e STMRunner) Run(txs [][]byte) ([]*abci.ExecTxResult, error) {
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
		memTxs, estimates = preEstimates(txs, e.workers, authStore, bankStore, e.coinDenom, e.txDecoder)
	}

	if err := ExecuteBlockWithEstimates(
		e.ctx,
		blockSize,
		index,
		e.ms,
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
			results[txn] = e.deliverTx(int(txn), memTx, ms, cache)

			if v != nil {
				incarnationCache[txn].Store(v)
			}
		},
	); err != nil {
		return nil, err
	}

	return results, nil
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
