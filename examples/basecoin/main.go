package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
	acm "github.com/cosmos/cosmos-sdk/x/account"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/sendtx"
)

func main() {

	app := app.NewApp("basecoin")

	db, err := dbm.NewGoLevelDB("basecoin", "basecoin-data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create CommitStoreLoader
	cacheSize := 10000
	numHistory := int64(100)
	loader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)

	// Create MultiStore
	multiStore := store.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader("main", loader)

	// Create Handler
	handler := types.ChainDecorators(
		// recover.Decorator(),
		// logger.Decorator(),
		auth.DecoratorFn(acm.NewAccountStore),
	).WithHandler(
		sendtx.TransferHandlerFn(acm.NewAccountStore),
	)

	// TODO: load genesis
	// TODO: InitChain with validators
	// accounts := acm.NewAccountStore(multiStore.GetKVStore("main"))
	// TODO: set the genesis accounts

	// Set everything on the app and load latest
	app.SetCommitMultiStore(multiStore)
	app.SetTxParser(txParser)
	app.SetHandler(handler)
	if err := app.LoadLatestVersion(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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
	return
}

func txParser(txBytes []byte) (types.Tx, error) {
	var tx sendtx.SendTx
	err := json.Unmarshal(txBytes, &tx)
	return tx, err
}

//-----------------------------------------------------------------------------
