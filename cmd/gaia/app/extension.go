package app

import (
	"database/sql"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *GaiaApp) BeginBlockHook(database *sql.DB, blockerFunctions []func(*GaiaApp, *sql.DB, sdk.Context, []sdk.ValAddress, []sdk.AccAddress, string, string, types.RequestBeginBlock), vals []sdk.ValAddress, accs []sdk.AccAddress, network string, chainid string) sdk.BeginBlocker {
	return func(ctx sdk.Context, req types.RequestBeginBlock) types.ResponseBeginBlock {
		res := app.BeginBlocker(ctx, req)
		// fucntions
		for _, fn := range blockerFunctions {
			fn(app, database, ctx, vals, accs, network, chainid, req)
		}
		return res
	}
}
