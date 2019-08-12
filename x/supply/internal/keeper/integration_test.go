package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/cosmos/cosmos-sdk/x/supply/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

var (
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
	holder     = "holder"
)

// nolint: deadcode unused
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app, ctx := simapp.Setup(isCheckTx)

	// add module accounts to supply keeper
	maccPerms := simapp.MaccPerms
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	app.SupplyKeeper = NewKeeper(app.Cdc, app.Keys[types.StoreKey], app.AccountKeeper, app.BankKeeper, maccPerms)
	app.SupplyKeeper.SetSupply(ctx, types.NewSupply(sdk.NewCoins()))

	return app, ctx
}
