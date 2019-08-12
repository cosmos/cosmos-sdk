package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (app *simapp.SimApp, ctx sdk.Context) {
	app, ctx = simapp.Setup(isCheckTx)
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}
