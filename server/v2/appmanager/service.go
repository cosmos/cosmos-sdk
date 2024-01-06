package appmanager

import (
	"context"
	"fmt"

	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf"
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

	mempool mempool.Mempool[T]

	initGenesis func(ctx context.Context, genesisBytes []byte) error

	prepareHandler appmanager.PrepareHandler[T]
	processHandler appmanager.ProcessHandler[T]

	stf *stf.STF[T]
}

// BuildBlock builds a block when requested by consensus. It will take in the total size txs to be included and return a list of transactions
func (a AppManager[T]) BuildBlock(ctx context.Context, height uint64, totalSize uint32) ([]T, error) {
	latestVersion, currentState, err := a.db.StateLatest()
	if err != nil {
		return nil, fmt.Errorf("unable to create new state for height %d: %w", height, err)
	}

	if latestVersion+1 != height {
		if err != nil {
			return nil, fmt.Errorf("invalid BuildBlock height wanted %d, got %d", latestVersion+1, height)
		}
	}

	txs, err := a.prepareHandler(ctx, totalSize, a.mempool, currentState)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (a AppManager[T]) VerifyBlock(ctx context.Context, height uint64, txs []T) error {
	latestVersion, currentState, err := a.db.StateLatest()
	if err != nil {
		return fmt.Errorf("unable to create new state for height %d: %w", height, err)
	}

	if latestVersion+1 != height {
		return fmt.Errorf("invalid VerifyBlock height wanted %d, got %d", latestVersion+1, height)
	}

	err = a.processHandler(ctx, txs, currentState)
	if err != nil {
		return err
	}

	return nil
}

func (a AppManager[T]) DeliverBlock(ctx context.Context, block appmanager.BlockRequest) (*appmanager.BlockResponse, []store.ChangeSet, error) {
	latestVersion, currentState, err := a.db.StateLatest()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create new state for height %d: %w", block.Height, err)
	}

	if latestVersion+1 != block.Height {
		return nil, nil, fmt.Errorf("invalid DeliverBlock height wanted %d, got %d", latestVersion+1, block.Height)
	}

	blockResponse, newState, err := a.stf.DeliverBlock(ctx, block, currentState)
	if err != nil {
		return nil, nil, fmt.Errorf("block delivery failed: %w", err)
	}

	newStateChanges, err := newState.ChangeSets()
	if err != nil {
		return nil, nil, fmt.Errorf("change set: %w", err)
	}

	return blockResponse, newStateChanges, nil
}

// CommitBlock commits the block to the database, it must be called after DeliverBlock or when Finalization criteria is met
func (a AppManager[T]) CommitBlock(ctx context.Context, height uint64, sc []store.ChangeSet) (Hash, error) {
	stateRoot, err := a.db.StateCommit(sc)
	if err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}
	return stateRoot, nil
}

func (a AppManager[T]) Simulate(ctx context.Context, tx []byte) (appmanager.TxResult, error) {
	_, state, err := a.db.StateLatest()
	if err != nil {
		return appmanager.TxResult{}, err
	}
	result := a.stf.Simulate(ctx, state, a.simulationGasLimit, tx)
	return result, nil
}

func (a AppManager[T]) Query(ctx context.Context, request Type) (response Type, err error) {
	_, queryState, err := a.db.StateLatest()
	if err != nil {
		return nil, err
	}
	return a.stf.Query(ctx, queryState, a.queryGasLimit, request)
}
