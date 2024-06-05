package storesource

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/indexer/base"
)

type source struct {
	engine *indexerbase.Engine
}

type Options struct {
	indexerbase.EngineOptions[appmodule.AppModule]
}

func NewSource(opts Options) storetypes.ABCIListener {
	engine := indexerbase.NewEngine(opts.EngineOptions)
	return &source{engine: engine}
}

func (s source) ListenFinalizeBlock(ctx context.Context, req abci.FinalizeBlockRequest, res abci.FinalizeBlockResponse) error {
	//TODO implement me
	// block
	// txs
	// events - missing event msg index
	return nil
}

func (s source) ListenCommit(ctx context.Context, res abci.CommitResponse, changeSet []*storetypes.StoreKVPair) error {
	for _, kv := range changeSet {
		if kv.Delete {
			if err := s.engine.ReceiveStateDelete(kv.StoreKey, kv.Key, false); err != nil {
				return err
			} else {
				if err := s.engine.ReceiveStateSet(kv.StoreKey, kv.Key, kv.Value); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return s.engine.Commit()
}

var _ storetypes.ABCIListener = &source{}
