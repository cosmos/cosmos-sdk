package app

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/cosmos/cosmos-sdk/examples/simpleGov/types"
	simpleGov "github.com/cosmos/cosmos-sdk/examples/simpleGov/x/simple_governance"

	tmtypes "github.com/tendermint/tendermint/types"
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
func NewSimpleGovApp(logger log.Logger, db dbm.DB, baseAppOptions ...func(*bam.BaseApp)) *SimpleGovApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &SimpleGovApp{
		BaseApp:              bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...),
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
		auth.ProtoBaseAccount,  // prototype
	)

	// Add handlers.
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.stakeKeeper = stake.NewKeeper(cdc, app.capKeyStakingStore, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))
	app.simpleGovKeeper = simpleGov.NewKeeper(cdc, app.capKeySimpleGovStore, app.coinKeeper, app.stakeKeeper, app.RegisterCodespace(simpleGov.DefaultCodespace))
	app.Router().
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		AddRoute("simpleGov", simpleGov.NewHandler(app.simpleGovKeeper))

	// perform initialization logic
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

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
	simpleGov.RegisterWire(cdc)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&types.AppAccount{}, "simpleGov/Account", nil)
	return cdc
}

// BeginBlocker reflects logic to run before any TXs application are processed
// by the application.
func (app *SimpleGovApp) BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return abci.ResponseBeginBlock{}
}

// EndBlocker reflects logic to run after all TXs are processed by the
// application.
func (app *SimpleGovApp) EndBlocker(_ sdk.Context, _ abci.RequestEndBlock) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{}
}

// initChainer implements the custom application logic that the BaseApp will
// invoke upon initialization. In this case, it will take the application's
// state provided by 'req' and attempt to deserialize said state. The state
// should contain all the genesis accounts. These accounts will be added to the
// application's account mapper.
func (app *SimpleGovApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(types.GenesisState)
	err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		// TODO: https://github.com/cosmos/cosmos-sdk/issues/468
		panic(err)
	}

	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAppAccount()
		if err != nil {
			// TODO: https://github.com/cosmos/cosmos-sdk/issues/468
			panic(err)
		}

		acc.AccountNumber = app.accountMapper.GetNextAccountNumber(ctx)
		app.accountMapper.SetAccount(ctx, acc)
	}

	return abci.ResponseInitChain{}
}

// ExportAppStateJSON handles the custom logic for state export
func (app *SimpleGovApp) ExportAppStateJSON() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
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
	appState, err = wire.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	return appState, validators, err
}
