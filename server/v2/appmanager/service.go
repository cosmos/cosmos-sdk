package appmanager

import (
	"context"
	"fmt"
	"sync/atomic"

	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/mempool"
	"cosmossdk.io/server/v2/stf"
)

// AppManager is a coordinator for all things related to an application
type AppManager[T transaction.Tx] struct {
	// configs
	checkTxGasLimit    uint64
	queryGasLimit      uint64
	simulationGasLimit uint64
	// configs - end

	db store.Store

	lastBlockHeight *atomic.Uint64

	initGenesis func(ctx context.Context, genesisBytes []byte) error

	stf *stf.STF[T]

	mempool mempool.Mempool[T]

	// txDecode defines a closure that is able to decode a tx into bytes.
	txDecode func(txBytes []byte) (T, error)
}

// InsertTx will attempt to insert a tx into the mempool in case it passes the validation steps.
// An error in appmanager.TxResult.Error means the tx was invalid.
// Otherwise, if an error is returned by the function itself, it means that something related
// to the state machine failed.
func (a AppManager[T]) InsertTx(ctx context.Context, txBytes []byte) (appmanager.TxResult, error) {
	decodedTx, err := a.txDecode(txBytes)
	if err != nil {
		return appmanager.TxResult{Error: err}, nil
	}
	validateState, err := a.getLatestState(ctx)
	if err != nil {
		return appmanager.TxResult{}, err
	}

	validationResult := a.stf.ValidateTx(ctx, validateState, a.checkTxGasLimit, decodedTx)
	// in case we error just return.
	if validationResult.Error != nil {
		return validationResult, nil
	}

	// otherwise insert into mempool, we also provided encoded tx bytes
	// because during block building the consensus engine will expect the Tx
	// into its raw bytes format.
	err = a.mempool.Push(ctx, mempool.NewCacheTx(decodedTx, txBytes))
	if err != nil {
		return appmanager.TxResult{}, fmt.Errorf("unable to push tx to mempool: %w", err)
	}
	return validationResult, nil
}

// DeliverBlock executes the block proposal.
func (a AppManager[T]) DeliverBlock(ctx context.Context, block appmanager.BlockRequest) (*appmanager.BlockResponse, Hash, error) {
	currentState, err := a.db.NewStateAt(block.Height)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create new state for height %d: %w", block.Height, err)
	}

	stfBlock, err := intoSTFBlock(a.txDecode, block)
	if err != nil {
		return nil, nil, err
	}
	blockResponse, newState, err := a.stf.DeliverBlock(ctx, stfBlock, currentState)
	if err != nil {
		return nil, nil, fmt.Errorf("block delivery failed: %w", err)
	}
	// apply new state to store
	newStateChanges, err := newState.ChangeSets()
	if err != nil {
		return nil, nil, fmt.Errorf("change set: %w", err)
	}
	stateRoot, err := a.db.CommitState(newStateChanges)
	if err != nil {
		return nil, nil, fmt.Errorf("commit failed: %w", err)
	}
	// update last stored block
	a.lastBlockHeight.Store(block.Height)
	return blockResponse, stateRoot, nil
}

// Simulate simulates the Tx over the last committed state.
func (a AppManager[T]) Simulate(ctx context.Context, txBytes []byte) (appmanager.TxResult, error) {
	// decode tx
	tx, err := a.txDecode(txBytes)
	if err != nil {
		return appmanager.TxResult{}, fmt.Errorf("unable to decode tx: %w", err)
	}
	// check if tx gas limit is in bounds
	if a.simulationGasLimit < tx.GetGasLimit() {
		return appmanager.TxResult{}, fmt.Errorf("simulated tx gas limit is higher than allowed: %d -> %d", a.simulationGasLimit, tx.GetGasLimit())
	}
	state, err := a.getLatestState(ctx)
	if err != nil {
		return appmanager.TxResult{}, err
	}

	result := a.stf.Simulate(ctx, state, tx)
	return result, nil
}

func (a AppManager[T]) Query(ctx context.Context, request Type) (response Type, err error) {
	queryState, err := a.getLatestState(ctx)
	if err != nil {
		return nil, err
	}
	return a.stf.Query(ctx, queryState, a.queryGasLimit, request)
}

// getLatestState provides a readonly view of the state of the last committed block.
func (a AppManager[T]) getLatestState(_ context.Context) (store.ReadonlyState, error) {
	lastBlock := a.lastBlockHeight.Load()
	lastBlockState, err := a.db.ReadonlyStateAt(lastBlock)
	if err != nil {
		return nil, err
	}
	return lastBlockState, nil
}

// intoSTFBlock uses the tx decoder to decode tx bytes into an stf.BlockRequest.
// NOTE: we assume consensus is not providing us invalid txs. It really should not,
// since all txs have passed through InsertTx which tests them.
// If this is not the case then we might need to merge results, from invalid un-decodable txs.
func intoSTFBlock[T transaction.Tx](txDecode func([]byte) (T, error), request appmanager.BlockRequest) (stf.BlockRequest[T], error) {
	txs := make([]T, len(request.Txs))
	for i, txBytes := range request.Txs {
		tx, err := txDecode(txBytes)
		if err != nil {
			return stf.BlockRequest[T]{}, fmt.Errorf("consensus provide an invalid tx in the block request, index %d: %w", i, err)
		}
		txs[i] = tx
	}
	return stf.BlockRequest[T]{
		Height:            request.Height,
		Time:              request.Time,
		Hash:              request.Hash,
		Txs:               txs,
		ConsensusMessages: request.ConsensusMessages,
	}, nil
}
