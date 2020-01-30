package app

import (
	"github.com/ChorusOne/hipparchuslibs/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *GaiaApp) BeginBlockHook(blockerFunctions []func(common.App, sdk.Context, types.RequestBeginBlock)) sdk.BeginBlocker {
	return func(ctx sdk.Context, req types.RequestBeginBlock) types.ResponseBeginBlock {
		res := app.BeginBlocker(ctx, req)
		ctx = ctx.
			WithGasMeter(sdk.NewInfiniteGasMeter()).
			WithBlockGasMeter(sdk.NewInfiniteGasMeter())
		// fucntions
		for _, fn := range blockerFunctions {
			fn(app, ctx, req)
		}
		return res
	}
}
