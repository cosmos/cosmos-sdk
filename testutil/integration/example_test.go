package integration_test

import (
	"fmt"
	"io"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/google/go-cmp/cmp"
)

// Example shows how to use the integration test framework to test the integration of SDK modules.
// Panics are used in this example, but in a real test case, you should use the testing.T object and assertions.
func Example() {
	// in this example we are testing the integration of the following modules:
	// - mint, which directly depends on auth, bank and staking
	encodingCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, mint.AppModuleBasic{})
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, minttypes.StoreKey)
	authority := authtypes.NewModuleAddress("gov").String()

	accountKeeper := authkeeper.NewAccountKeeper(
		encodingCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		"cosmos",
		authority,
	)

	// subspace is nil because we don't test params (which is legacy anyway)
	authModule := auth.NewAppModule(encodingCfg.Codec, accountKeeper, authsims.RandomGenesisAccounts, nil)

	// here bankkeeper and staking keeper is nil because we are not testing them
	// subspace is nil because we don't test params (which is legacy anyway)
	mintKeeper := mintkeeper.NewKeeper(encodingCfg.Codec, keys[minttypes.StoreKey], nil, accountKeeper, nil, authtypes.FeeCollectorName, authority)
	mintModule := mint.NewAppModule(encodingCfg.Codec, mintKeeper, accountKeeper, nil, nil)

	// create the application and register all the modules from the previous step
	// replace the name and the logger by testing values in a real test case (e.g. t.Name() and log.NewTestLogger(t))
	integrationApp := integration.NewIntegrationApp("example", log.NewLogger(io.Discard), keys, authModule, mintModule)

	// register the message and query servers
	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(accountKeeper))
	minttypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), mintkeeper.NewMsgServerImpl(mintKeeper))
	minttypes.RegisterQueryServer(integrationApp.QueryHelper(), mintKeeper)

	params := minttypes.DefaultParams()
	params.BlocksPerYear = 10000

	// now we can use the application to test a mint message
	result, err := integrationApp.RunMsg(&minttypes.MsgUpdateParams{
		Authority: authority,
		Params:    params,
	})
	if err != nil {
		panic(err)
	}

	// in this example the result is an empty response, a nil check is enough
	// in other cases, it is recommended to check the result value.
	if result == nil {
		panic(fmt.Errorf("unexpected nil result"))
	}

	// we now check the result
	resp := minttypes.MsgUpdateParamsResponse{}
	err = encodingCfg.Codec.Unmarshal(result.Value, &resp)
	if err != nil {
		panic(err)
	}

	// we should also check the state of the application
	got := mintKeeper.GetParams(integrationApp.SDKContext())
	if diff := cmp.Diff(got, params); diff != "" {
		panic(diff)
	}
	fmt.Println(got.BlocksPerYear)
	// Output: 10000
}

// ExampleOneModule shows how to use the integration test framework to test the integration of a single module.
// That module has no dependency on other modules.
func Example_oneModule() {
	// in this example we are testing the integration of the auth module:
	encodingCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey)
	authority := authtypes.NewModuleAddress("gov").String()

	accountKeeper := authkeeper.NewAccountKeeper(
		encodingCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		"cosmos",
		authority,
	)

	// subspace is nil because we don't test params (which is legacy anyway)
	authModule := auth.NewAppModule(encodingCfg.Codec, accountKeeper, authsims.RandomGenesisAccounts, nil)

	// create the application and register all the modules from the previous step
	// replace the name and the logger by testing values in a real test case (e.g. t.Name() and log.NewTestLogger(t))
	integrationApp := integration.NewIntegrationApp("example-one-module", log.NewLogger(io.Discard), keys, authModule)

	// register the message and query servers
	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(accountKeeper))

	params := authtypes.DefaultParams()
	params.MaxMemoCharacters = 1000

	// now we can use the application to test a mint message
	result, err := integrationApp.RunMsg(&authtypes.MsgUpdateParams{
		Authority: authority,
		Params:    params,
	})
	if err != nil {
		panic(err)
	}

	// in this example the result is an empty response, a nil check is enough
	// in other cases, it is recommended to check the result value.
	if result == nil {
		panic(fmt.Errorf("unexpected nil result"))
	}

	// we now check the result
	resp := authtypes.MsgUpdateParamsResponse{}
	err = encodingCfg.Codec.Unmarshal(result.Value, &resp)
	if err != nil {
		panic(err)
	}

	// we should also check the state of the application
	got := accountKeeper.GetParams(integrationApp.SDKContext())
	if diff := cmp.Diff(got, params); diff != "" {
		panic(diff)
	}
	fmt.Println(got.MaxMemoCharacters)
	// Output: 1000
}
