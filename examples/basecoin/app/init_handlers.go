package app

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// initCapKeys, initBaseApp, initStores, initHandlers.
func (app *BasecoinApp) initHandlers() {
	app.initDefaultAnteHandler()
	app.initRouterHandlers()
}

func (app *BasecoinApp) initDefaultAnteHandler() {
	var authAnteHandler = auth.NewAnteHandler(app.accountMapper)
	app.BaseApp.SetDefaultAnteHandler(authAnteHandler)
}

func (app *BasecoinApp) initRouterHandlers() {
	var router = app.BaseApp.Router()
	var accountMapper = app.accountMapper

	// All handlers must be added here.
	// The order matters.
	router.AddRoute("bank", bank.NewHandler(accountMapper))
}
