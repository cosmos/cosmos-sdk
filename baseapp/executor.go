package baseapp

import (
	"context"
	"cosmossdk.io/collections"
	"cosmossdk.io/store/cachemulti"
	storetypes "cosmossdk.io/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"io"
	"sync"
	"sync/atomic"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	blockstm "github.com/crypto-org-chain/go-block-stm"
)

// Executor the interface for implementing custom execution logic, such as block-stm
type Executor interface {
	run(txs [][]byte) ([]*abci.ExecTxResult, error)
}

// DefaultExecutor default executor without parallelism
type DefaultExecutor struct {
	ctx       context.Context
	txDecoder sdk.TxDecoder
	deliverTx func(tx []byte) *abci.ExecTxResult
}

func (d DefaultExecutor) run(txs [][]byte) ([]*abci.ExecTxResult, error) {
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

// STMExecutor simple implementation of block-stm
type STMExecutor struct {
	ctx       context.Context
	txDecoder sdk.TxDecoder
	stores    []storetypes.StoreKey
	ms        storetypes.MultiStore
	workers   int
	estimate  bool
	coinDenom string
	deliverTx func(int, sdk.Tx, storetypes.MultiStore, map[string]any) *abci.ExecTxResult
}

func (e STMExecutor) run(txs [][]byte) ([]*abci.ExecTxResult, error) {
	var authStore, bankStore int
	index := make(map[storetypes.StoreKey]int, len(e.stores))
	for i, k := range e.stores {
		switch k.Name() {
		case authtypes.StoreKey:
			authStore = i
		case banktypes.StoreKey:
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
		estimates []blockstm.MultiLocations
		memTxs    []sdk.Tx
	)

	if e.estimate {
		memTxs, estimates = preEstimates(txs, e.workers, authStore, bankStore, e.coinDenom, e.txDecoder)
	}

	if err := blockstm.ExecuteBlockWithEstimates(
		e.ctx,
		blockSize,
		index,
		stmMultiStoreWrapper{e.ms},
		e.workers,
		estimates,
		func(txn blockstm.TxnIndex, ms blockstm.MultiStore) {
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
			results[txn] = e.deliverTx(int(txn), memTx, msWrapper{ms}, cache)

			if v != nil {
				incarnationCache[txn].Store(v)
			}
		},
	); err != nil {
		return nil, err
	}

	return results, nil

}

type msWrapper struct {
	blockstm.MultiStore
}

func (ms msWrapper) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	//TODO implement me
	panic("implement me")
}

func (ms msWrapper) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	//TODO implement me
	panic("implement me")
}

func (ms msWrapper) LatestVersion() int64 {
	//TODO implement me
	panic("implement me")
}

var _ storetypes.MultiStore = msWrapper{}

func (ms msWrapper) getCacheWrapper(key storetypes.StoreKey) storetypes.CacheWrapper {
	return ms.GetStore(key)
}

func (ms msWrapper) GetStore(key storetypes.StoreKey) storetypes.Store {
	return ms.MultiStore.GetStore(key)
}

func (ms msWrapper) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return ms.MultiStore.GetKVStore(key)
}

func (ms msWrapper) GetObjKVStore(key storetypes.StoreKey) storetypes.ObjKVStore {
	return ms.MultiStore.GetObjKVStore(key)
}

func (ms msWrapper) CacheMultiStore() storetypes.CacheMultiStore {
	return cachemulti.NewFromParent(ms.getCacheWrapper, nil, nil)
}

// CacheWrap Implements CacheWrapper.
func (ms msWrapper) CacheWrap() storetypes.CacheWrap {
	return ms.CacheMultiStore().(storetypes.CacheWrap)
}

// GetStoreType returns the type of the store.
func (ms msWrapper) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

// SetTracer Implements interface MultiStore
func (ms msWrapper) SetTracer(io.Writer) storetypes.MultiStore {
	return nil
}

// SetTracingContext Implements interface MultiStore
func (ms msWrapper) SetTracingContext(storetypes.TraceContext) storetypes.MultiStore {
	return nil
}

// TracingEnabled Implements interface MultiStore
func (ms msWrapper) TracingEnabled() bool {
	return false
}

type stmMultiStoreWrapper struct {
	storetypes.MultiStore
}

var _ blockstm.MultiStore = stmMultiStoreWrapper{}

func (ms stmMultiStoreWrapper) GetStore(key storetypes.StoreKey) storetypes.Store {
	return ms.MultiStore.GetStore(key)
}

func (ms stmMultiStoreWrapper) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return ms.MultiStore.GetKVStore(key)
}

// preEstimates returns a static estimation of the written keys for each transaction.
// NOTE: make sure it sync with the latest sdk logic when sdk upgrade.
func preEstimates(txs [][]byte, workers, authStore, bankStore int, coinDenom string, txDecoder sdk.TxDecoder) ([]sdk.Tx, []blockstm.MultiLocations) {
	memTxs := make([]sdk.Tx, len(txs))
	estimates := make([]blockstm.MultiLocations, len(txs))

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
				authtypes.AddressStoreKeyPrefix,
				sdk.AccAddressKey,
				feePayer,
			)
			if err != nil {
				continue
			}

			// balance key
			balanceKey, err := collections.EncodeKeyWithPrefix(
				banktypes.BalancesPrefix,
				collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
				collections.Join(feePayer, coinDenom),
			)
			if err != nil {
				continue
			}

			estimates[i] = blockstm.MultiLocations{
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
