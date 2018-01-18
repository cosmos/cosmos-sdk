package app

import (
	"fmt"
	"os"

	apm "github.com/cosmos/cosmos-sdk/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/abci/server"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
)

const appName = "BasecoinApp"

type BasecoinApp struct {
	*apm.App
	cdc        *wire.Codec
	multiStore sdk.CommitMultiStore

	// The key to access the substores.
	mainStoreKey *sdk.KVStoreKey
	ibcStoreKey  *sdk.KVStoreKey

	// Additional stores:
	accStore sdk.AccountStore
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

	// TODO: Load genesis
	// TODO: InitChain with validators
	// TODO: Set the genesis accounts
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

// depends on initMultiStore()
func (app *BasecoinApp) initSDKApp() {
	app.App = apm.NewApp(appName, app.multiStore)
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
	if err := app.LoadLatestVersion(app.mainStoreKey); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
