package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/abci/server"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	bcm "github.com/cosmos/cosmos-sdk/examples/basecoin/types"
)

func main() {

	// First, create the Application.
	app := sdk.NewApp("basecoin")

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

	// The key to access the main KVStore.
	var mainStoreKey = new(KVStoreKey)
	var ibcStoreKey = new(KVStoreKey)

	// Create MultiStore
	multiStore := store.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader("main", mainStoreKey, mainLoader)
	multiStore.SetSubstoreLoader("ibc", ibcStoreKey, ibcLoader)
	app.SetCommitMultiStore(multiStore)

	// Set Tx decoder
	app.SetTxDecoder(decodeTx)

	var accStore = auth.NewAccountStore(mainStoreKey, bcm.AppAccountCodec{})
	var authAnteHandler = auth.NewAnteHandler(accStore)

	// Handle charging fees and checking signatures.
	app.SetDefaultAnteHandler(authAnteHandler)

	// Add routes to App.
	app.Router().AddRoute("bank", bank.NewHandler(accStore))

	// TODO: load genesis
	// TODO: InitChain with validators
	// accounts := auth.NewAccountStore(multiStore.GetKVStore("main"))
	// TODO: set the genesis accounts

	// Load the stores.
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

func registerMsgs() {
	wire.RegisterInterface((*types.Msg), nil)
	wire.RegisterConcrete((*bank.SendMsg), nil)
}

func decodeTx(txBytes []byte) (types.Tx, error) {
	var tx = sdk.StdTx{}
	err := wire.UnmarshalBinary(txBytes, &tx)
	return tx, err
}
