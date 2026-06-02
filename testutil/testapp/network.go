package testapp

import (
	"os"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// SDKAppFixture returns a network.TestFixtureFactory that constructs an SDKApp
// for in-process network tests. Use it instead of network.DefaultConfigWithAppConfig.
//
// Example:
//
//	cfg := network.DefaultConfig(testapp.SDKAppFixture)
func SDKAppFixture() network.TestFixture {
	tempDir, err := os.MkdirTemp("", "testapp-fixture-")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	opts := simtestutil.AppOptionsMap{
		flags.FlagHome:    tempDir,
		flags.FlagChainID: "test-chain",
	}

	cfg := app.DefaultSDKAppConfig("app", opts)
	sdkApp := app.NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	sdkApp.LoadModules()

	if err := sdkApp.LoadLatestVersion(); err != nil {
		panic("failed to load latest version: " + err.Error())
	}

	encCfg := moduletestutil.TestEncodingConfig{
		InterfaceRegistry: sdkApp.InterfaceRegistry(),
		Codec:             sdkApp.AppCodec(),
		TxConfig:          sdkApp.TxConfig(),
		Amino:             sdkApp.LegacyAmino(),
	}

	return network.TestFixture{
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			home := val.GetCtx().Config.RootDir
			minGasPrices := val.GetAppConfig().MinGasPrices
			pruning := val.GetAppConfig().Pruning

			appOpts := simtestutil.AppOptionsMap{
				flags.FlagHome:    home,
				flags.FlagChainID: val.GetCtx().Viper.GetString(flags.FlagChainID),
			}

			appCfg := app.DefaultSDKAppConfig(
				"app", appOpts,
				baseapp.SetMinGasPrices(minGasPrices),
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(pruning)),
			)

			newApp := app.NewSDKApp(val.GetCtx().Logger, dbm.NewMemDB(), nil, appCfg)
			newApp.LoadModules()
			if err := newApp.LoadLatestVersion(); err != nil {
				panic("failed to load latest version: " + err.Error())
			}
			return newApp
		},
		GenesisState:   sdkApp.DefaultGenesis(),
		EncodingConfig: encCfg,
	}
}
