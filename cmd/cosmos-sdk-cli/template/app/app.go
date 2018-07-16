package app

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/cosmos-sdk/cmd/cosmos-sdk-cli/template/types"
)

const (
	appName = "MyAwesomeProjectApp"
)

// Extended ABCI application
type MyAwesomeProjectApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore    *sdk.KVStoreKey
	capKeyAccountStore *sdk.KVStoreKey

	// keepers
	feeCollectionKeeper auth.FeeCollectionKeeper
	coinKeeper          bank.Keeper

	// Manage getting and setting accounts
	accountMapper auth.AccountMapper
}

func NewMyAwesomeProjectApp(logger log.Logger, db dbm.DB) *MyAwesomeProjectApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &MyAwesomeProjectApp{
		BaseApp:            bam.NewBaseApp(appName, cdc, logger, db),
		cdc:                cdc,
		capKeyMainStore:    sdk.NewKVStoreKey("main"),
		capKeyAccountStore: sdk.NewKVStoreKey("acc"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.capKeyAccountStore,  // target store
		types.ProtoAppAccount(), // prototype
	)

	// Add handlers.
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.Router().
		AddRoute("bank", bank.NewHandler(app.coinKeeper))

	// Initialize BaseApp.
	app.SetInitChainer(app.initChainerFn())
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
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
	bank.RegisterWire(cdc)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&types.AppAccount{}, "myawesomeproject/Account", nil)

	cdc.Seal()

	return cdc
}

// custom logic for myawesomeproject initialization
// nolint: unparam
func (app *MyAwesomeProjectApp) initChainerFn() sdk.InitChainer {
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

		return abci.ResponseInitChain{}
	}
}

// Custom logic for state export
func (app *MyAwesomeProjectApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
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
	return appState, validators, nil
}
