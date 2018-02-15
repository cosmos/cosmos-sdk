package app

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/sketchy"
)

// initCapKeys, initBaseApp, initStores, initHandlers.
func (app *BasecoinApp) initHandlers() {
	app.initRouterHandlers()
	app.initAnteHandlers()
}

func (app *BasecoinApp) initAnteHandlers() {

	// Deducts fee from payer.
	// Verifies signatures and nonces.
	// Sets Signers to ctx.
	app.BaseApp.SetDefaultAnteHandler(
		auth.NewAnteHandler(app.accountMapper))


	// init custom ante handlers
	// Example: app.router.AddAnte("paymentchannels", paymentchannels.NewAnteHandler(app.accountMapper))
}

func (app *BasecoinApp) initRouterHandlers() {

	// All handlers must be added here.
	// The order matters.
	app.router.AddRoute("bank", bank.NewHandler(bank.NewCoinKeeper(app.accountMapper)))
	app.router.AddRoute("sketchy", sketchy.NewHandler())
}
