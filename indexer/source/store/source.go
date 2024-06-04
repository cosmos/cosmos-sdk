package store

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/indexer"
)

type Source struct {
	engine indexer.Engine
}

func (s Source) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	//TODO implement me
	// block
	// txs
	// events - missing event msg index
	return nil
}

func (s Source) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
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

var _ storetypes.ABCIListener = &Source{}
