package simapp

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/app"
	countertypes "github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v053 to v054.
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.53.x to v0.54.x.
const UpgradeName = "v053-to-v054"

var MyUpgrade = app.Upgrade[*SimApp]{
	Name: UpgradeName,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			countertypes.ModuleName,
		},
	},
	UpgradeCallBack: func(ctx sdk.Context, plan upgradetypes.Plan, app *SimApp) error {
		ctx.Logger().Debug("this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
		return nil
	},
}
