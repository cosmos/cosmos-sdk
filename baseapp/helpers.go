package baseapp

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (app *BaseApp) Check(tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	return app.runTx(runTxModeCheck, nil, tx)
}

func (app *BaseApp) Simulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, sdk.Tx, error) {
	tx, err := app.txDecoder(txBytes)
	if err != nil {
		return sdk.GasInfo{}, nil, nil, sdkerrors.Wrap(err, "failed to decode tx")
	}
	gi, r, err := app.runTx(runTxModeSimulate, txBytes, tx)
	return gi, r, tx, err
}

func (app *BaseApp) Deliver(tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	return app.runTx(runTxModeDeliver, nil, tx)
}

// Context with current {check, deliver}State of the app used by tests.
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.checkState.ms, header, true, app.logger).
			WithMinGasPrices(app.minGasPrices)
	}

	return sdk.NewContext(app.deliverState.ms, header, false, app.logger)
}

func (app *BaseApp) NewUncachedContext(isCheckTx bool, header abci.Header) sdk.Context {
	return sdk.NewContext(app.cms, header, isCheckTx, app.logger)
}
