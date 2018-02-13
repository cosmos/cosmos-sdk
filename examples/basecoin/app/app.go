package app

import (
	"fmt"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/abci/server"
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

// TODO: This should take in more configuration options.
// TODO: This should be moved into baseapp to isolate complexity
func NewBasecoinApp(genesisPath string) *BasecoinApp {

	// Create and configure app.
	var app = &BasecoinApp{}

	// TODO open up out of functions, or introduce clarity,
	// interdependancies are a nightmare to debug
	app.initCapKeys() // ./init_capkeys.go
	app.initBaseApp() // ./init_baseapp.go
	app.initStores()  // ./init_stores.go
	app.initBaseAppInitStater()
	app.initHandlers() // ./init_handlers.go

	genesisiDoc, err := bam.GenesisDocFromFile(genesisPath)
	if err != nil {
		panic(fmt.Errorf("error loading genesis state: %v", err))
	}

	// set up the cache store for ctx, get ctx
	// TODO: can InitChain handle this too ?
	app.BaseApp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{}})
	ctx := app.BaseApp.NewContext(false, nil) // context for DeliverTx

	// TODO: combine with InitChain and let tendermint invoke it.
	err = app.BaseApp.InitStater(ctx, genesisiDoc.AppState)
	if err != nil {
		panic(fmt.Errorf("error initializing application genesis state: %v", err))
	}

	app.loadStores()

	return app
}

// RunForever - BasecoinApp execution and cleanup
func (app *BasecoinApp) RunForever() {

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:46658", "socket", app)
	if err != nil {
		cmn.Exit(err.Error())
	}
	srv.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})

}

// Load the stores
func (app *BasecoinApp) loadStores() {
	if err := app.LoadLatestVersion(app.capKeyMainStore); err != nil {
		cmn.Exit(err.Error())
	}
}
