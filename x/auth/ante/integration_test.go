package ante_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/KiraCore/cosmos-sdk/simapp"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	authtypes "github.com/KiraCore/cosmos-sdk/x/auth/types"
)

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}
