package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// initCapKeys() happens before initStores(), initSDKApp(), and initRoutes().
func (app *BasecoinApp) initCapKeys() {

	// All top-level capabilities keys
	// should be constructed here.
	// For more information, see http://www.erights.org/elib/capability/ode/ode.pdf.
	app.mainStoreKey = sdk.NewKVStoreKey("main")
	app.ibcStoreKey = sdk.NewKVStoreKey("ibc")

}
