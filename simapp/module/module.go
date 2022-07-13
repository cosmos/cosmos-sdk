package module

import (
	dbm "github.com/tendermint/tm-db"

	simappv1alpha1 "cosmossdk.io/api/cosmos/simapp/module/v1alpha1"
	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	networktypes "github.com/cosmos/cosmos-sdk/types/module/testutil/network"
)

func init() {
	appmodule.Register(&simappv1alpha1.Module{},
		appmodule.Provide(provideModule))
}

func provideModule() (networktypes.AppConstructor,
	networktypes.GenesisState,
	codectypes.InterfaceRegistry,
	codec.Codec,
	*codec.LegacyAmino,
	client.TxConfig) {
	appCtr := func(val networktypes.Validator) servertypes.Application {
		return simapp.NewSimApp(
			val.GetCtx().Logger, dbm.NewMemDB(), nil, true,
			simapp.MakeTestEncodingConfig(),
			simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir),
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
		)
	}
	e := simapp.MakeTestEncodingConfig()
	gs := simapp.ModuleBasics.DefaultGenesis(e.Codec)

	return appCtr, gs, e.InterfaceRegistry, e.Codec, e.Amino, e.TxConfig
}
