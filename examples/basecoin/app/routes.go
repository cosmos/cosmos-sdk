package app

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Handle charging tx fees and checking signatures.
func (app *BasecoinApp) initAnteHandler() {
	var authAnteHandler = auth.NewAnteHandler(app.accStore)
	app.App.SetDefaultAnteHandler(authAnteHandler)
}

// Constructs router to route handling of msgs.
func (app *BasecoinApp) initRoutes() {
	var router = app.App.Router()
	// var multiStore = app.multiStore
	var accStore = app.accStore

	router.AddRoute("bank", bank.NewHandler(accStore))
	// more routes here... (order matters)
}
