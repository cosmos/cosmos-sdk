package appmanager

import (
	"context"
	"fmt"
	"sync/atomic"
)

type Mempool interface {
}

type AppManagerBuilder struct {
	InitGenesis map[string]func(ctx context.Context, moduleGenesisBytes []byte) error
}

func (a *AppManagerBuilder) RegisterInitGenesis(moduleName string, genesisFunc func(ctx context.Context, moduleGenesisBytes []byte) error) {
	a.InitGenesis[moduleName] = genesisFunc
}

func (a *AppManagerBuilder) RegisterHandler(moduleName, handlerName string, handler MsgHandler) {
	panic("...")
}

type MsgSetKVPairs struct {
	Pairs []ChangeSet
}

func (a *AppManagerBuilder) Build() *AppManager {
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
	return &AppManager{initGenesis: genesis}
}

type AppManager struct {
	// configs
	checkTxGasLimit uint64
	queryGasLimit   uint64
	// configs - end

	lastBlock *atomic.Uint64

	initGenesis func(ctx context.Context, genesisBytes []byte) error

	stf *STFAppManager

	mempool Mempool
}

func (a AppManager) CheckTx(ctx context.Context, txBytes []byte) error {
	// decode tx
	tx, err := a.stf.txDecoder.Decode(txBytes)
	if err != nil {
		return err
	}
	// apply validation using last block state
	checkTxState := a.getLatestState(ctx)
	_, _, err = a.stf.validateTx(ctx, checkTxState, min(a.checkTxGasLimit, tx.GetGasLimit()), tx)
	if err != nil {
		return err
	}
	// TODO: cache, insert in mempool?
	return nil
}

func (a AppManager) DeliverBlock(ctx context.Context, block Block) (*BlockResponse, error) {
	blockResponse, err := a.stf.DeliverBlock(ctx, block)
	if err != nil {
		return nil, err
	}
	// update last stored block
	a.lastBlock.Store(block.Height)
	return blockResponse, nil
}

func (a AppManager) Query(ctx context.Context, request Type) (response Type, err error) {
	queryState := a.getLatestState(ctx)
	queryCtx := a.stf.makeContext(ctx, queryState, a.queryGasLimit)
	return a.stf.queryRouter.Handle(queryCtx, request)
}

// getLatestState provides a readonly view of the state of the last committed block.
func (a AppManager) getLatestState(_ context.Context) BranchStore {
	lastBlock := a.lastBlock.Load()
	lastBlockStore := a.stf.store.ReadonlyWithVersion(lastBlock)
	return a.stf.branch(lastBlockStore)
}
