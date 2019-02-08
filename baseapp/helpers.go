package baseapp

import (
	"regexp"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// nolint - Mostly for testing
func (app *BaseApp) Check(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(runTxModeCheck, nil, tx)
}

// nolint - full tx execution
func (app *BaseApp) Simulate(txBytes []byte, tx sdk.Tx) (result sdk.Result) {
	return app.runTx(runTxModeSimulate, txBytes, tx)
}

// nolint
func (app *BaseApp) Deliver(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(runTxModeDeliver, nil, tx)
}

// Context with current {check, deliver}State of the app
// used by tests
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.checkState.ms, header, true, app.logger).
			WithMinGasPrices(app.minGasPrices)
	}

	return sdk.NewContext(app.deliverState.ms, header, false, app.logger)
}
