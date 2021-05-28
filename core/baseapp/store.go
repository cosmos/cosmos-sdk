package baseapp

import (
	"context"
	"crypto/sha256"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

type StoreMiddleware struct {
	cms       sdk.CommitMultiStore
	deliverMs types.CacheMultiStore
	checkMs   types.CacheMultiStore
}

var _ ABCIConsensusMiddleware = &StoreMiddleware{}
var _ ABCIMempoolMiddleware = &StoreMiddleware{}

func (s *StoreMiddleware) OnInitChain(ctx context.Context, req abci.RequestInitChain, next ABCIConsensusHandler) abci.ResponseInitChain {
	if req.InitialHeight > 1 {
		err := s.cms.SetInitialVersion(req.InitialHeight)
		if err != nil {
			panic(err)
		}
	}

	s.deliverMs = s.cms.CacheMultiStore()
	s.checkMs = s.cms.CacheMultiStore()

	// TODO add ms to ctx

	res := next.InitChain(ctx, req)

	if lastCommitId := s.cms.LastCommitID(); !lastCommitId.IsZero() {
		res.AppHash = lastCommitId.Hash
	} else {
		// $ echo -n '' | sha256sum
		// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		emptyHash := sha256.Sum256([]byte{})
		res.AppHash = emptyHash[:]
	}

	return res
}

func (s StoreMiddleware) OnBeginBlock(ctx context.Context, req abci.RequestBeginBlock, next ABCIConsensusHandler) abci.ResponseBeginBlock {
	if s.cms.TracingEnabled() {
		s.cms.SetTracingContext(sdk.TraceContext(
			map[string]interface{}{"blockHeight": req.Header.Height},
		))
	}

	if s.deliverMs == nil {
		s.deliverMs = s.cms.CacheMultiStore()
	}

	return next.BeginBlock(ctx, req)
}

func (s StoreMiddleware) CheckTx(ctx context.Context, tx abci.RequestCheckTx, handler ABCIMempoolHandler) abci.ResponseCheckTx {
	panic("implement me")
}

func (s StoreMiddleware) OnDeliverTx(ctx context.Context, tx abci.RequestDeliverTx, handler ABCIConsensusHandler) abci.ResponseDeliverTx {
	panic("implement me")
}

func (s StoreMiddleware) OnEndBlock(ctx context.Context, block abci.RequestEndBlock, handler ABCIConsensusHandler) abci.ResponseEndBlock {
	panic("implement me")
}

func (s StoreMiddleware) OnCommit(ctx context.Context, next ABCIConsensusHandler) abci.ResponseCommit {
	// Write the DeliverTx state into branched storage and commit the MultiStore.
	// The write to the DeliverTx state writes all state transitions to the root
	// MultiStore (app.cms) so when Commit() is called is persists those values.
	s.deliverMs.Write()
	commitID := s.cms.Commit()
	//app.logger.Info("commit synced", "commit", fmt.Sprintf("%X", commitID))

	// Reset the Check state to the latest committed.
	//
	// NOTE: This is safe because Tendermint holds a lock on the mempool for
	// Commit. Use the header from this latest block.
	s.checkMs = s.cms.CacheMultiStore()

	// empty/reset the deliver state
	s.deliverMs = nil

	res := next.Commit(ctx)

	res.Data = commitID.Hash

	return res
}
