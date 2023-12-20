package appmanager

import (
	"context"
	"fmt"
	"sync/atomic"

	"cosmossdk.io/server/v2/stf"

	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
)

type AppManagerBuilder[T transaction.Tx] struct {
	InitGenesis map[string]func(ctx context.Context, moduleGenesisBytes []byte) error
}

func (a *AppManagerBuilder[T]) RegisterInitGenesis(moduleName string, genesisFunc func(ctx context.Context, moduleGenesisBytes []byte) error) {
	a.InitGenesis[moduleName] = genesisFunc
}

func (a *AppManagerBuilder[T]) RegisterHandler(moduleName, handlerName string, handler stf.MsgHandler) {
	panic("...")
}

type MsgSetKVPairs struct {
	Pairs []store.ChangeSet
}

func (a *AppManagerBuilder[T]) Build() *AppManager[T] {
	genesis := func(ctx context.Context, genesisBytes []byte) error {
		genesisMap := map[string][]byte{} // module=> genesis bytes
		for module, genesisFunc := range a.InitGenesis {
			err := genesisFunc(ctx, genesisMap[module])
			if err != nil {
				return fmt.Errorf("failed to init genesis on module: %s", module)
			}
		}
		return nil
	}
	return &AppManager[T]{initGenesis: genesis}
}

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
}

func (a AppManager[T]) DeliverBlock(ctx context.Context, block appmanager.BlockRequest) (*appmanager.BlockResponse, Hash, error) {
	currentState, err := a.db.NewStateAt(block.Height)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create new state for height %d: %w", block.Height, err)
	}

	blockResponse, newState, err := a.stf.DeliverBlock(ctx, block, currentState)
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

func (a AppManager[T]) Simulate(ctx context.Context, tx []byte) (appmanager.TxResult, error) {
	state, err := a.getLatestState(ctx)
	if err != nil {
		return appmanager.TxResult{}, err
	}
	result := a.stf.Simulate(ctx, state, a.simulationGasLimit, tx)
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
