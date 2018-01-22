package app

import (
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// initCapKeys, initBaseApp, initStores, initRoutes.
func (app *BasecoinApp) initStores() {
	app.mountStores()
	app.initAccountMapper()
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
