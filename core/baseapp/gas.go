package baseapp

import (
	"context"

	"github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetConsensusParams(ctx context.Context) types.ConsensusParams {
	panic("TODO")
}

func getMaximumBlockGas(ctx context.Context) uint64 {
	panic("TODO")
}

type gasMiddleware struct{}

type gasKeyTy string

const gasKey gasKeyTy = "gas"

func (g gasMiddleware) CheckTx(ctx context.Context, tx types.RequestCheckTx, handler ABCIMempoolHandler) types.ResponseCheckTx {
	panic("implement me")
}

func (g gasMiddleware) OnInitChain(ctx context.Context, chain types.RequestInitChain, handler ABCIConsensusHandler) types.ResponseInitChain {
	panic("implement me")
}

func (g gasMiddleware) OnBeginBlock(ctx context.Context, block types.RequestBeginBlock, handler ABCIConsensusHandler) types.ResponseBeginBlock {
	var gasMeter sdk.GasMeter
	if maxGas := getMaximumBlockGas(ctx); maxGas > 0 {
		gasMeter = sdk.NewGasMeter(maxGas)
	} else {
		gasMeter = sdk.NewInfiniteGasMeter()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockGasMeter(gasMeter)

	ctx = context.WithValue(ctx, sdk.SdkContextKey, sdkCtx)

	return handler.BeginBlock(ctx, block)
}

func (g gasMiddleware) OnDeliverTx(ctx context.Context, tx types.RequestDeliverTx, handler ABCIConsensusHandler) types.ResponseDeliverTx {
	panic("implement me")
}

func (g gasMiddleware) OnEndBlock(ctx context.Context, block types.RequestEndBlock, handler ABCIConsensusHandler) types.ResponseEndBlock {
	panic("implement me")
}

func (g gasMiddleware) OnCommit(ctx context.Context, handler ABCIConsensusHandler) types.ResponseCommit {
	panic("implement me")
}

var _ ABCIConsensusMiddleware = gasMiddleware{}
var _ ABCIMempoolMiddleware = gasMiddleware{}
