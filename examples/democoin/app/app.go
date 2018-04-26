package app

import (
	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"

	"github.com/cosmos/cosmos-sdk/examples/democoin/types"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/pow"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/simplestake"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/sketchy"
)

const (
	appName = "DemocoinApp"
)

// Extended ABCI application
type DemocoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore    *sdk.KVStoreKey
	capKeyAccountStore *sdk.KVStoreKey
	capKeyPowStore     *sdk.KVStoreKey
	capKeyIBCStore     *sdk.KVStoreKey
	capKeyStakingStore *sdk.KVStoreKey

	// keepers
	powKeeper pow.Keeper

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
}

func NewDemocoinApp(logger log.Logger, db dbm.DB) *DemocoinApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &DemocoinApp{
		BaseApp:            bam.NewBaseApp(appName, cdc, logger, db),
		cdc:                cdc,
		capKeyMainStore:    sdk.NewKVStoreKey("main"),
		capKeyAccountStore: sdk.NewKVStoreKey("acc"),
		capKeyPowStore:     sdk.NewKVStoreKey("pow"),
		capKeyIBCStore:     sdk.NewKVStoreKey("ibc"),
		capKeyStakingStore: sdk.NewKVStoreKey("stake"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	)

	// Add handlers.
	coinKeeper := bank.NewKeeper(app.accountMapper)
	coolKeeper := cool.NewKeeper(app.capKeyMainStore, coinKeeper, app.RegisterCodespace(cool.DefaultCodespace))
	app.powKeeper = pow.NewKeeper(app.capKeyPowStore, pow.NewConfig("pow", int64(1)), coinKeeper, app.RegisterCodespace(pow.DefaultCodespace))
	ibcMapper := ibc.NewMapper(app.cdc, app.capKeyIBCStore, app.RegisterCodespace(ibc.DefaultCodespace))
	stakeKeeper := simplestake.NewKeeper(app.capKeyStakingStore, coinKeeper, app.RegisterCodespace(simplestake.DefaultCodespace))
	app.Router().
		AddRoute("bank", bank.NewHandler(coinKeeper)).
		AddRoute("cool", cool.NewHandler(coolKeeper)).
		AddRoute("pow", app.powKeeper.Handler).
		AddRoute("sketchy", sketchy.NewHandler()).
		AddRoute("ibc", ibc.NewHandler(ibcMapper, coinKeeper)).
		AddRoute("simplestake", simplestake.NewHandler(stakeKeeper))

	// Initialize BaseApp.
	app.SetInitChainer(app.initChainerFn(coolKeeper, app.powKeeper))
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeyPowStore, app.capKeyIBCStore, app.capKeyStakingStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, auth.BurnFeeHandler))
	err := app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc) // Register crypto.
	sdk.RegisterWire(cdc)    // Register Msgs
	cool.RegisterWire(cdc)
	pow.RegisterWire(cdc)
	bank.RegisterWire(cdc)
	ibc.RegisterWire(cdc)
	simplestake.RegisterWire(cdc)

	// Register AppAccount
	cdc.RegisterInterface((*sdk.Account)(nil), nil)
	cdc.RegisterConcrete(&types.AppAccount{}, "democoin/Account", nil)
	return cdc
}

// custom logic for democoin initialization
func (app *DemocoinApp) initChainerFn(coolKeeper cool.Keeper, powKeeper pow.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		stateJSON := req.AppStateBytes

		genesisState := new(types.GenesisState)
		err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			// return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.ToAppAccount()
			if err != nil {
				panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
				//	return sdk.ErrGenesisParse("").TraceCause(err, "")
			}
			app.accountMapper.SetAccount(ctx, acc)
		}

		// Application specific genesis handling
		err = coolKeeper.InitGenesis(ctx, genesisState.CoolGenesis)
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		err = powKeeper.InitGenesis(ctx, genesisState.POWGenesis)
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		return abci.ResponseInitChain{}
	}
}

// Custom logic for state export
func (app *DemocoinApp) ExportGenesis() types.GenesisState {
	ctx := app.NewContext(true, abci.Header{})
	return types.GenesisState{
		Accounts:    []*types.GenesisAccount{},
		POWGenesis:  app.powKeeper.WriteGenesis(ctx),
		CoolGenesis: cool.Genesis{},
	}
}
