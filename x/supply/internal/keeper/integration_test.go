package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/supply/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

var (
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
	holder     = "holder"
)

// nolint:deadcode,unused
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)

	// add module accounts to supply keeper
	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.SupplyKeeper = keep.NewKeeper(app.Codec(), app.GetKey(types.StoreKey), app.AccountKeeper, app.BankKeeper, maccPerms)
	app.SupplyKeeper.SetSupply(ctx, types.NewSupply(sdk.NewCoins()))

	return app, ctx
}
