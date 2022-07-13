package simapp

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
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func init() {
	appmodule.Register(&simappv1alpha1.Module{},
		appmodule.Provide(provideModule))
}

func provideModule() (network.AppConstructor,
	network.GenesisState,
	codectypes.InterfaceRegistry,
	codec.Codec,
	*codec.LegacyAmino,
	client.TxConfig) {
	appCtr := func(val network.Validator) servertypes.Application {
		return NewSimApp(
			val.Ctx.Logger, dbm.NewMemDB(), nil, true,
			MakeTestEncodingConfig(),
			simtestutil.NewAppOptionsWithFlagHome(val.Ctx.Config.RootDir),
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
	e := MakeTestEncodingConfig()
	gs := ModuleBasics.DefaultGenesis(e.Codec)

	return appCtr, gs, e.InterfaceRegistry, e.Codec, e.Amino, e.TxConfig
}
