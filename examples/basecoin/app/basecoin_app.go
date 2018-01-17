package app

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/tendermint/abci/server"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
)

const appName = "BasecoinApp"

type BasecoinApp struct {
	*sdk.App
	cdc        *wire.Codec
	multiStore sdk.CommitMultiStore
	appStore   *auth.AccountStore

	// The key to access the main KVStore.
	mainStoreKey *sdk.KVStoreKey
	ibcStoreKey  *sdk.KVStoreKey
}

// TODO: This should take in more configuration options.
func NewBasecoinApp() *BasecoinApp {

	// Create and configure app.
	var app = &BasecoinApp{}
	app.initKeys()
	app.initMultiStore()
	app.initAppStore()
	app.initSDKApp()
	app.initCodec()
	app.initTxDecoder()
	app.initAnteHandler()
	app.initRoutes()

	// TODO: load genesis
	// TODO: InitChain with validators
	// accounts := auth.NewAccountStore(multiStore.GetKVStore("main"))
	// TODO: set the genesis accounts
	app.loadStores()

	return app
}

func (app *BasecoinApp) RunForever() {

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:46658", "socket", app)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	srv.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})

}

func (app *BasecoinApp) initKeys() {
	app.mainStoreKey = sdk.NewKVStoreKey("main")
	app.ibcStoreKey = sdk.NewKVStoreKey("ibc")
}

// depends on initKeys()
func (app *BasecoinApp) initMultiStore() {

	// Create the underlying leveldb datastore which will
	// persist the Merkle tree inner & leaf nodes.
	db, err := dbm.NewGoLevelDB("basecoin", "basecoin-data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create CommitStoreLoader.
	cacheSize := 10000
	numHistory := int64(100)
	mainLoader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)
	ibcLoader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)

	// Create MultiStore
	multiStore := store.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader(app.mainStoreKey, mainLoader)
	multiStore.SetSubstoreLoader(app.ibcStoreKey, ibcLoader)

	// Finally,
	app.multiStore = multiStore
}

// depends on initKeys()
func (app *BasecoinApp) initAppStore() {
	app.appStore = auth.NewAccountStore(
		app.mainStoreKey,
		types.AppAccountCodec{},
	)
}

// depends on initMultiStore()
func (app *BasecoinApp) initSDKApp() {
	app.App = sdk.NewApp(appName, app.multiStore)
}

func (app *BasecoinApp) initCodec() {
	app.cdc = wire.NewCodec()
}

// depends on initSDKApp()
func (app *BasecoinApp) initTxDecoder() {
	app.App.SetTxDecoder(app.decodeTx)
}

// initAnteHandler defined in app/routes.go
// initRoutes defined in app/routes.go

// Load the stores.
func (app *BasecoinApp) loadStores() {
	if err := app.LoadLatestVersion(mainStoreKey); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
