package app

import (
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// initCapKeys, initBaseApp, initStores, initRoutes.
func (app *BasecoinApp) initRoutes() {
	var router = app.BaseApp.Router()
	var accountMapper = app.accountMapper

	// All handlers must be added here.
	// The order matters.
	router.AddRoute("bank", bank.NewHandler(accountMapper))
}
