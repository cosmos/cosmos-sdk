package baseapp

import sdk "github.com/cosmos/cosmos-sdk/types"

// Returns a new Context suitable for AnteHandler (and indirectly Handler) processing.
// NOTE: txBytes may be nil to support TestApp.RunCheckTx
// and TestApp.RunDeliverTx.
func (app *BaseApp) newContext(isCheckTx bool, txBytes []byte) sdk.Context {
	var store sdk.MultiStore
	if isCheckTx {
		store = app.msCheck
	} else {
		store = app.msDeliver
	}
	if store == nil {
		panic("BaseApp.NewContext() requires BeginBlock(): missing store")
	}
	if app.header == nil {
		panic("BaseApp.NewContext() requires BeginBlock(): missing header")
	}

	// Initialize arguments to Handler.
	var ctx = sdk.NewContext(
		store,
		*app.header,
		isCheckTx,
		txBytes,
	)
	return ctx
}
