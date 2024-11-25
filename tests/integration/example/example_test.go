package integration_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/mint"
	mintkeeper "cosmossdk.io/x/mint/keeper"
	minttypes "cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Example shows how to use the integration test framework to test the integration of SDK modules.
// Panics are used in this example, but in a real test case, you should use the testing.T object and assertions.
func Example() {
	// in this example we are testing the integration of the following modules:
	// - mint, which directly depends on auth, bank and staking
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, mint.AppModule{})
	signingCtx := encodingCfg.InterfaceRegistry.SigningContext()

	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, minttypes.StoreKey)
	authority := authtypes.NewModuleAddress("gov").String()

	// replace the logger by testing values in a real test case (e.g. log.NewTestLogger(t))
	logger := log.NewNopLogger()

	// gomock initializations
	ctrl := gomock.NewController(&testing.T{})
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)
	accNum := uint64(0)
	acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (uint64, error) {
		currentNum := accNum
		accNum++
		return currentNum, nil
	})

	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		encodingCfg.Codec,
		authtypes.ProtoBaseAccount,
		acctsModKeeper,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		addresscodec.NewBech32Codec("cosmos"),
		"cosmos",
		authority,
	)

	// subspace is nil because we don't test params (which is legacy anyway)
	authModule := auth.NewAppModule(encodingCfg.Codec, accountKeeper, acctsModKeeper, authsims.RandomGenesisAccounts, nil)

	// here bankkeeper and staking keeper is nil because we are not testing them
	// subspace is nil because we don't test params (which is legacy anyway)
	mintKeeper := mintkeeper.NewKeeper(encodingCfg.Codec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[minttypes.StoreKey]), logger), accountKeeper, nil, authtypes.FeeCollectorName, authority)
	mintModule := mint.NewAppModule(encodingCfg.Codec, mintKeeper, accountKeeper)

	// create the application and register all the modules from the previous step
	integrationApp := integration.NewIntegrationApp(
		logger,
		keys,
		encodingCfg.Codec,
		signingCtx.AddressCodec(),
		signingCtx.ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			authtypes.ModuleName: authModule,
			minttypes.ModuleName: mintModule,
		},
		baseapp.NewMsgServiceRouter(),
		baseapp.NewGRPCQueryRouter(),
	)

	// register the message and query servers
	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(accountKeeper))
	minttypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), mintkeeper.NewMsgServerImpl(mintKeeper))
	minttypes.RegisterQueryServer(integrationApp.QueryHelper(), mintkeeper.NewQueryServerImpl(mintKeeper))

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
		panic(errors.New("unexpected nil result"))
	}

	// we now check the result
	resp := minttypes.MsgUpdateParamsResponse{}
	err = encodingCfg.Codec.Unmarshal(result.Value, &resp)
	if err != nil {
		panic(err)
	}

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// we should also check the state of the application
	got, err := mintKeeper.Params.Get(sdkCtx)
	if err != nil {
		panic(err)
	}

	if diff := cmp.Diff(got, params); diff != "" {
		panic(diff)
	}
	fmt.Println(got.BlocksPerYear)
	// Output: 10000
}

// Example_oneModule shows how to use the integration test framework to test the integration of a single module.
// That module has no dependency on other modules.
func Example_oneModule() {
	// in this example we are testing the integration of the auth module:
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{})
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey)
	authority := authtypes.NewModuleAddress("gov").String()

	// replace the logger by testing values in a real test case (e.g. log.NewTestLogger(t))
	logger := log.NewLogger(io.Discard)

	// gomock initializations
	ctrl := gomock.NewController(&testing.T{})
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)
	accNum := uint64(0)
	acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (uint64, error) {
		currentNum := accNum
		accNum++
		return currentNum, nil
	})

	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		encodingCfg.Codec,
		authtypes.ProtoBaseAccount,
		acctsModKeeper,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		addresscodec.NewBech32Codec("cosmos"),
		"cosmos",
		authority,
	)

	// subspace is nil because we don't test params (which is legacy anyway)
	authModule := auth.NewAppModule(encodingCfg.Codec, accountKeeper, acctsModKeeper, authsims.RandomGenesisAccounts, nil)

	// create the application and register all the modules from the previous step
	integrationApp := integration.NewIntegrationApp(
		logger,
		keys,
		encodingCfg.Codec,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			authtypes.ModuleName: authModule,
		},
		baseapp.NewMsgServiceRouter(),
		baseapp.NewGRPCQueryRouter(),
	)

	// register the message and query servers
	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(accountKeeper))

	params := authtypes.DefaultParams()
	params.MaxMemoCharacters = 1000

	// now we can use the application to test a mint message
	result, err := integrationApp.RunMsg(&authtypes.MsgUpdateParams{
		Authority: authority,
		Params:    params,
	},
		// this allows to the begin and end blocker of the module before and after the message
		integration.WithAutomaticFinalizeBlock(),
		// this allows to commit the state after the message
		integration.WithAutomaticCommit(),
	)
	if err != nil {
		panic(err)
	}

	// verify that the begin and end blocker were called
	// NOTE: in this example, we are testing auth, which doesn't have any begin or end blocker
	// so verifying the block height is enough
	if integrationApp.LastBlockHeight() != 2 {
		panic(fmt.Errorf("expected block height to be 2, got %d", integrationApp.LastBlockHeight()))
	}

	// in this example the result is an empty response, a nil check is enough
	// in other cases, it is recommended to check the result value.
	if result == nil {
		panic(errors.New("unexpected nil result"))
	}

	// we now check the result
	resp := authtypes.MsgUpdateParamsResponse{}
	err = encodingCfg.Codec.Unmarshal(result.Value, &resp)
	if err != nil {
		panic(err)
	}

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// we should also check the state of the application
	got := accountKeeper.GetParams(sdkCtx)
	if diff := cmp.Diff(got, params); diff != "" {
		panic(diff)
	}
	fmt.Println(got.MaxMemoCharacters)
	// Output: 1000
}
