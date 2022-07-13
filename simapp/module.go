package simapp

import (
	dbm "github.com/tendermint/tm-db"

	simappv1alpha1 "cosmossdk.io/api/cosmos/simapp/module/v1alpha1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/baseapp"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func init() {
	appmodule.Register(&simappv1alpha1.Module{},
		appmodule.Provide(provideModule))
}

type simappOut struct {
	depinject.Out
	params.EncodingConfig
	AppConstructor network.AppConstructor
}

func provideModule() simappOut {
	appCtr := func(val network.Validator) servertypes.Application {
		return NewSimApp(
			val.Ctx.Logger, dbm.NewMemDB(), nil, true,
			MakeTestEncodingConfig(),
			simtestutil.NewAppOptionsWithFlagHome(val.Ctx.Config.RootDir),
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
	encCfg := MakeTestEncodingConfig()

	s := simappOut{AppConstructor: appCtr}
	s.InterfaceRegistry = encCfg.InterfaceRegistry
	s.Codec = encCfg.Codec
	s.Amino = encCfg.Amino
	s.TxConfig = encCfg.TxConfig

	return s
}
