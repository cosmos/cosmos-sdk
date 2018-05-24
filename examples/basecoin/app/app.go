package app

import (
	"encoding/json"

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
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
)

const (
	appName = "BasecoinApp"
)

// Extended ABCI application
type BasecoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyIBC     *sdk.KVStoreKey
	keyStake   *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper auth.AccountMapper
	coinKeeper    bank.Keeper
	ibcMapper     ibc.Mapper
	stakeKeeper   stake.Keeper
}

func NewBasecoinApp(logger log.Logger, db dbm.DB) *BasecoinApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &BasecoinApp{
		BaseApp:    bam.NewBaseApp(appName, cdc, logger, db),
		cdc:        cdc,
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		keyIBC:     sdk.NewKVStoreKey("ibc"),
		keyStake:   sdk.NewKVStoreKey("stake"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.keyAccount,      // target store
		&types.AppAccount{}, // prototype
	)

	// add accountMapper/handlers
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.ibcMapper = ibc.NewMapper(app.cdc, app.keyIBC, app.RegisterCodespace(ibc.DefaultCodespace))
	app.stakeKeeper = stake.NewKeeper(app.cdc, app.keyStake, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))

	// register message routes
	app.Router().
		AddRoute("auth", auth.NewHandler(app.accountMapper)).
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper))

	// Initialize BaseApp.
	app.SetInitChainer(app.initChainer)
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper))
	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// Custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc) // Register crypto.
	sdk.RegisterWire(cdc)    // Register Msgs
	bank.RegisterWire(cdc)
	stake.RegisterWire(cdc)
	ibc.RegisterWire(cdc)

	// register custom AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&types.AppAccount{}, "basecoin/Account", nil)
	return cdc
}

// Custom logic for basecoin initialization
func (app *BasecoinApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
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

// Custom logic for state export
func (app *BasecoinApp) ExportAppStateJSON() (appState json.RawMessage, err error) {
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
