package app

import (
	"fmt"
	"os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/abci/server"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
)

const appName = "BasecoinApp"

// BasecoinApp - extended ABCI application
type BasecoinApp struct {
	*bam.BaseApp
	router     bam.Router
	cdc        *wire.Codec
	multiStore sdk.CommitMultiStore //TODO distinguish this store from *bam.BaseApp.cms <- is this one master?? confused

	// The key to access the substores.
	capKeyMainStore *sdk.KVStoreKey
	capKeyIBCStore  *sdk.KVStoreKey

	// Object mappers:
	accountMapper sdk.AccountMapper
}

// NewBasecoinApp - create new BasecoinApp
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

// Load the stores
func (app *BasecoinApp) loadStores() {
	if err := app.LoadLatestVersion(app.capKeyMainStore); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
