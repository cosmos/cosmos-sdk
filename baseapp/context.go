package baseapp

import sdk "github.com/cosmos/cosmos-sdk/types"

// NOTE: Unstable.
// Returns a new Context suitable for AnteHandler (and indirectly Handler) processing.
func (app *BaseApp) NewContext(isCheckTx bool, txBytes []byte) sdk.Context {
	var store sdk.MultiStore
	if isCheckTx {
		store = app.msCheck
	} else {
		store = app.msDeliver
	}

	// Initialize arguments to Handler.
	var ctx = sdk.NewContext(
		store,
		app.header,
		isCheckTx,
		txBytes,
	)
	return ctx
}
