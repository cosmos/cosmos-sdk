// Package keeper provides methods to initialize SDK keepers with local storage for test purposes
package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	feemarketkeeper "cosmossdk.io/x/feemarket/keeper"
	feemarkettypes "cosmossdk.io/x/feemarket/types"
)

const govModuleName = "gov"

// TestKeepers holds all keepers used during keeper tests for all modules
type TestKeepers struct {
	FeeMarketKeeper *feemarketkeeper.Keeper
}

// TestMsgServers holds all message servers used during keeper tests for all modules
type TestMsgServers struct {
	FeeMarketMsgServer feemarkettypes.MsgServer
}

var additionalMaccPerms = map[string][]string{
	feemarkettypes.ModuleName:       nil,
	feemarkettypes.FeeCollectorName: {authtypes.Burner},
}

// NewTestSetup returns initialized instances of all the keepers and message servers of the modules
func NewTestSetup(t testing.TB) (sdk.Context, TestKeepers, TestMsgServers) {
	// options = append(options, testkeeper.WithAdditionalModuleAccounts(additionalMaccPerms))

	// _, tk, tms := testkeeper.NewTestSetup(t, options...)

	// // initialize extra keeper
	// feeMarketKeeper := FeeMarket(tk.Initializer, tk.AccountKeeper)
	// require.NoError(t, tk.Initializer.LoadLatest())

	// // initialize msg servers
	// feeMarketMsgSrv := feemarketkeeper.NewMsgServer(feeMarketKeeper)

	// ctx := sdk.NewContext(tk.Initializer.StateStore, tmproto.Header{
	// 	Time:   testkeeper.ExampleTimestamp,
	// 	Height: testkeeper.ExampleHeight,
	// }, false, log.NewNopLogger())

	// err := feeMarketKeeper.SetState(ctx, feemarkettypes.DefaultState())
	// require.NoError(t, err)
	// err = feeMarketKeeper.SetParams(ctx, feemarkettypes.DefaultParams())
	// require.NoError(t, err)

	// testKeepers := TestKeepers{
	// 	TestKeepers:     tk,
	// 	FeeMarketKeeper: feeMarketKeeper,
	// }

	// testMsgServers := TestMsgServers{
	// 	TestMsgServers:     tms,
	// 	FeeMarketMsgServer: feeMarketMsgSrv,
	// }

	// return ctx, testKeepers, testMsgServers
	return nil, TestKeepers{}, TestMsgServers{}
}

// FeeMarket initializes the fee market module using the testkeepers intializer.
// func FeeMarket(
// 	initializer any, // TODO: fix
// 	authKeeper authkeeper.AccountKeeper,
// ) *feemarketkeeper.Keeper {
// 	storeKey := storetypes.NewKVStoreKey(feemarkettypes.StoreKey)
// 	// initializer.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, initializer.DB)

// 	return feemarketkeeper.NewKeeper(
// 		initializer.Codec,
// 		storeKey,
// 		authKeeper,
// 		&feemarkettypes.TestDenomResolver{},
// 		authtypes.NewModuleAddress(govModuleName).String(),
// 	)
// }
