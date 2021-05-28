package baseapp

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HeaderMiddleware struct {
	initialHeight int64
}

var _ ABCIConsensusMiddleware = &HeaderMiddleware{}
var _ ABCIMempoolMiddleware = &HeaderMiddleware{}

func (h *HeaderMiddleware) OnInitChain(ctx context.Context, req abci.RequestInitChain, next ABCIConsensusHandler) abci.ResponseInitChain {
	// On a new chain, we consider the init chain block height as 0, even though
	// req.InitialHeight is 1 by default.
	initHeader := tmproto.Header{ChainID: req.ChainId, Time: req.Time}

	// If req.InitialHeight is > 1, then we set the initial version in the
	// stores.
	if req.InitialHeight > 1 {
		h.initialHeight = req.InitialHeight
		initHeader = tmproto.Header{ChainID: req.ChainId, Height: req.InitialHeight, Time: req.Time}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(initHeader)
	ctx = context.WithValue(ctx, sdk.SdkContextKey, sdkCtx)

	return next.InitChain(ctx, req)
}

func (h HeaderMiddleware) CheckTx(ctx context.Context, tx abci.RequestCheckTx, handler ABCIMempoolHandler) abci.ResponseCheckTx {
	panic("implement me")
}

func (h HeaderMiddleware) validateHeight(req abci.RequestBeginBlock) error {
	panic("TODO")
}

func (h HeaderMiddleware) OnBeginBlock(ctx context.Context, req abci.RequestBeginBlock, next ABCIConsensusHandler) abci.ResponseBeginBlock {
	if err := h.validateHeight(req); err != nil {
		panic(err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.
		WithBlockHeader(req.Header).
		WithBlockHeight(req.Header.Height)
	ctx = context.WithValue(ctx, sdk.SdkContextKey, sdkCtx)

	return next.BeginBlock(ctx, req)
}

func (h HeaderMiddleware) OnDeliverTx(ctx context.Context, tx abci.RequestDeliverTx, handler ABCIConsensusHandler) abci.ResponseDeliverTx {
	panic("implement me")
}

func (h HeaderMiddleware) OnEndBlock(ctx context.Context, block abci.RequestEndBlock, handler ABCIConsensusHandler) abci.ResponseEndBlock {
	panic("implement me")
}

func (h HeaderMiddleware) OnCommit(ctx context.Context, handler ABCIConsensusHandler) abci.ResponseCommit {
	//header := sdk.UnwrapSDKContext(ctx).BlockHeader()
	//retainHeight := app.GetBlockRetentionHeight(header.Height)
	//
	//// Write the DeliverTx state into branched storage and commit the MultiStore.
	//// The write to the DeliverTx state writes all state transitions to the root
	//// MultiStore (app.cms) so when Commit() is called is persists those values.
	//app.deliverState.ms.Write()
	//commitID := app.cms.Commit()
	//app.logger.Info("commit synced", "commit", fmt.Sprintf("%X", commitID))
	//
	//// Reset the Check state to the latest committed.
	////
	//// NOTE: This is safe because Tendermint holds a lock on the mempool for
	//// Commit. Use the header from this latest block.
	//app.setCheckState(header)
	//
	//// empty/reset the deliver state
	//app.deliverState = nil
	//
	//var halt bool
	//
	//switch {
	//case app.haltHeight > 0 && uint64(header.Height) >= app.haltHeight:
	//	halt = true
	//
	//case app.haltTime > 0 && header.Time.Unix() >= int64(app.haltTime):
	//	halt = true
	//}
	//
	//if halt {
	//	// Halt the binary and allow Tendermint to receive the ResponseCommit
	//	// response with the commit ID hash. This will allow the node to successfully
	//	// restart and process blocks assuming the halt configuration has been
	//	// reset or moved to a more distant value.
	//	app.halt()
	//}
	//
	//if app.snapshotInterval > 0 && uint64(header.Height)%app.snapshotInterval == 0 {
	//	go app.snapshot(header.Height)
	//}
	//
	//return abci.ResponseCommit{
	//	Data:         commitID.Hash,
	//	RetainHeight: retainHeight,
	//}
	panic("TODO")
}
