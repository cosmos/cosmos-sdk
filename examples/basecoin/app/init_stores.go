package app

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth"
	dbm "github.com/tendermint/tmlibs/db"
)

// initStores() happens after initCapKeys(), but before initSDKApp() and initRoutes().
func (app *BasecoinApp) initStores() {
	app.initMultiStore()
	app.initAccountStore()
}

// Initialize root MultiStore.
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

// Initialize the AccountStore, which accesses the MultiStore.
func (app *BasecoinApp) initAccountStore() {
	accStore := auth.NewAccountStore(
		// where accounts are persisted in the MultiStore.
		app.mainStoreKey,
		// prototype sdk.Account.
		&types.AppAccount{},
	)

	// If there are additional interfaces & concrete types that
	// need to be registered w/ wire.Codec, they can be registered
	// here before the accStore is sealed.
	cdc := accStore.WireCodec()
	auth.RegisterWireBaseAccount(cdc)

	app.accStore = accStore.Seal()
}
