package app

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth"
	dbm "github.com/tendermint/tmlibs/db"
)

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

	// Finally, set variables on app
	app.multiStore = multiStore
	app.storeKeys = map[string]store.SubstoreKey{}
	app.storeKeys[app.mainStoreKey.Name()] = app.mainStoreKey
	app.storeKeys[app.ibcStoreKey.Name()] = app.ibcStoreKey
}

// depends on initKeys()
func (app *BasecoinApp) initAppStore() {
	app.accStore = auth.NewAccountStore(
		app.mainStoreKey,
		types.AppAccountCodec{},
	)
}
