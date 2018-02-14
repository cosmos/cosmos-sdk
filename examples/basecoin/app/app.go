package app

import (
	"encoding/json"
	"fmt"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/sketchy"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
)

const appName = "BasecoinApp"

// Extended ABCI application
type BasecoinApp struct {
	*bam.BaseApp
	router bam.Router
	cdc    *wire.Codec

	// keys to access the substores
	capKeyMainStore *sdk.KVStoreKey
	capKeyIBCStore  *sdk.KVStoreKey

	// object mappers
	accountMapper sdk.AccountMapper
}

// construct top level keys
func (app *BasecoinApp) initCapKeys() {
	app.capKeyMainStore = sdk.NewKVStoreKey("main")
	app.capKeyIBCStore = sdk.NewKVStoreKey("ibc")
}

func (app *BasecoinApp) initDefaultAnteHandler() {
	// deducts fee from payer, verifies signatures and nonces, sets Signers to ctx.
	app.BaseApp.SetDefaultAnteHandler(auth.NewAnteHandler(app.accountMapper))
}

func (app *BasecoinApp) initRouterHandlers() {

	// All handlers must be added here, the order matters
	app.router.AddRoute("bank", bank.NewHandler(bank.NewCoinKeeper(app.accountMapper)))
	app.router.AddRoute("sketchy", sketchy.NewHandler())
}

func (app *BasecoinApp) initBaseAppTxDecoder() {
	cdc := makeTxCodec()
	app.BaseApp.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = sdk.StdTx{}
		// StdTx.Msg is an interface whose concrete
		// types are registered in app/msgs.go.
		err := cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxParse("").TraceCause(err, "")
		}
		return tx, nil
	})
}

// define the custom logic for basecoin initialization
func (app *BasecoinApp) initBaseAppInitStater() {
	accountMapper := app.accountMapper

	app.BaseApp.SetInitStater(func(ctx sdk.Context, state json.RawMessage) sdk.Error {
		if state == nil {
			return nil
		}

		genesisState := new(types.GenesisState)
		err := json.Unmarshal(state, genesisState)
		if err != nil {
			return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.ToAppAccount()
			if err != nil {
				return sdk.ErrGenesisParse("").TraceCause(err, "")
			}
			accountMapper.SetAccount(ctx, acc)
		}
		return nil
	})
}

// Initialize root stores.
func (app *BasecoinApp) mountStores() {

	// Create MultiStore mounts.
	app.BaseApp.MountStore(app.capKeyMainStore, sdk.StoreTypeIAVL)
	app.BaseApp.MountStore(app.capKeyIBCStore, sdk.StoreTypeIAVL)
}

// Initialize the AccountMapper.
func (app *BasecoinApp) initAccountMapper() {

	var accountMapper = auth.NewAccountMapper(
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	)

	// Register all interfaces and concrete types that
	// implement those interfaces, here.
	cdc := accountMapper.WireCodec()
	auth.RegisterWireBaseAccount(cdc)

	// Make accountMapper's WireCodec() inaccessible.
	app.accountMapper = accountMapper.Seal()
}

func NewBasecoinApp(genesisPath string) *BasecoinApp {

	// create and configure app
	var app = &BasecoinApp{}
	bapp := bam.NewBaseApp(appName)
	app.BaseApp = bapp
	app.router = bapp.Router()
	app.initBaseAppTxDecoder()

	// add keys
	app.initCapKeys()

	// initialize the stores
	app.mountStores()
	app.initAccountMapper()

	// initialize the genesis function
	app.initBaseAppInitStater()

	// initialize the handler
	app.initDefaultAnteHandler()
	app.initRouterHandlers()

	genesisAppState, err := bam.ReadGenesisAppState(genesisPath)
	if err != nil {
		panic(fmt.Errorf("error loading genesis state: %v", err))
	}

	// set up the cache store for ctx, get ctx
	// TODO: combine with InitChain and let tendermint invoke it.
	app.BaseApp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{}})
	ctx := app.BaseApp.NewContext(false, nil) // context for DeliverTx
	err = app.BaseApp.InitStater(ctx, genesisAppState)
	if err != nil {
		panic(fmt.Errorf("error initializing application genesis state: %v", err))
	}

	// load the stores
	if err := app.LoadLatestVersion(app.capKeyMainStore); err != nil {
		cmn.Exit(err.Error())
	}

	return app
}
