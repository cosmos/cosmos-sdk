package app

import (
	"encoding/json"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/simplestake"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/gamarin2/cosmos-sdk/examples/simpleGov/types"
	simpleGov "github.com/gamarin2/cosmos-sdk/examples/simpleGov/x/simple_governance"
	"github.com/gamarin2/cosmos-sdk/x/stake"
)

const (
	appName = "SimpleGovApp"
)

// SimpleGovApp extends Basecoin ABCI application
type SimpleGovApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore      *sdk.KVStoreKey
	capKeyAccountStore   *sdk.KVStoreKey
	capKeyStakingStore   *sdk.KVStoreKey
	capKeySimpleGovStore *sdk.KVStoreKey

	// keepers
	feeCollectionKeeper auth.FeeCollectionKeeper
	coinKeeper          bank.Keeper
	stakeKeeper         stake.Keeper
	simpleGovKeeper     simpleGov.Keeper

	// Manage getting and setting accounts
	accountMapper auth.AccountMapper
}

// NewSimpleGovApp creates a new SimpleGovApp instance
func NewSimpleGovApp(logger log.Logger, db dbm.DB) *SimpleGovApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &SimpleGovApp{
		BaseApp:              bam.NewBaseApp(appName, cdc, logger, db),
		cdc:                  cdc,
		capKeyMainStore:      sdk.NewKVStoreKey("main"),
		capKeyAccountStore:   sdk.NewKVStoreKey("acc"),
		capKeyStakingStore:   sdk.NewKVStoreKey("stake"),
		capKeySimpleGovStore: sdk.NewKVStoreKey("simpleGov"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.capKeyAccountStore, // target store
		&types.AppAccount{},    // prototype
	)

	// Add handlers.
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.stakeKeeper = stake.NewKeeper(app.capKeyStakingStore, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))
	app.simpleGovKeeper = simpleGov.NewKeeper(app.capKeySimpleGovStore, app.coinKeeper, app.stakeKeeper, app.RegisterCodespace(simpleGov.DefaultCodespace))
	app.Router().
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		AddRoute("simpleGov", simpleGov.NewHandler(app.simpleGovKeeper))

	// Initialize BaseApp.
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeySimpleGovStore, app.capKeyStakingStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
	err := app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// MakeCodec creates a custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc) // Register crypto.
	sdk.RegisterWire(cdc)    // Register Msgs
	bank.RegisterWire(cdc)
	simplestake.RegisterWire(cdc)
	simpleGov.RegisterWire(cdc)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&types.AppAccount{}, "simpleGov/Account", nil)
	return cdc
}

// ExportAppStateJSON handles the custom logic for state export
func (app *SimpleGovApp) ExportAppStateJSON() (appState json.RawMessage, err error) {
	ctx := app.NewContext(true, abci.Header{})

	// iterate to get the accounts
	accounts := []*types.GenesisAccount{}
	appendAccount := func(acc auth.Account) (stop bool) {
		account := &types.GenesisAccount{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}
		accounts = append(accounts, account)
		return false
	}
	app.accountMapper.IterateAccounts(ctx, appendAccount)

	genState := types.GenesisState{
		Accounts: accounts,
	}
	return wire.MarshalJSONIndent(app.cdc, genState)
}
