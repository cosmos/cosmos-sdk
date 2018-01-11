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
	"github.com/cosmos/cosmos-sdk/x/coinstore"
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
	mainLoader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)
	ibcLoader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)

	// The key to access the main KVStore.
	var mainKey = storeKey("main")
	var ibcKey = storeKey("ibc")

	// Create MultiStore
	multiStore := store.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader(mainKey, mainLoader)
	multiStore.SetSubstoreLoader(ibcKey, ibcLoader)

	// XXX
	var appAccountCodec AccountCodec = nil

	// Create Handler
	handler := types.ChainDecorators(
		recover.Decorator(),
		logger.Decorator(),
		auth.Decorator(appAccountCodec),
		fees.Decorator(mainKey),
		rollbackDecorator(),      // XXX define.
		ibc.Decorator(ibcKey),    // Handle IBC messages.
		pos.Decorator(mainKey),   // Handle staking messages.
		gov.Decorator(mainKey),   // Handle governance messages.
		coins.Decorator(mainKey), // Handle coinstore messages.
	).WithHandler(func(ctx types.context, tx Tx) Result {
		/*
			switch tx.(type) {
			case CustomTx1: ...
			case CustomTx2: ...
			}
		*/
	})

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

//----------------------------------------
// Misc.

func txParser(txBytes []byte) (types.Tx, error) {
	var tx coinstore.SendTx
	err := json.Unmarshal(txBytes, &tx)
	return tx, err
}

// an unexported (private) key which no module could know of unless
// it was passed in from the app.
type storeKey struct {
	writeable bool
	name      string
}

func newStoreKey(name string) storeKey {
	return storeKey{true, name}
}
func (s storeKey) ReadOnly() storeKey {
	return storeKey{false, s.name}
}
