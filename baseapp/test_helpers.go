package baseapp

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SimCheck defines a CheckTx helper function that used in tests and simulations.
func (app *BaseApp) SimCheck(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// runTx expects tx bytes as argument, so we encode the tx argument into
	// bytes. Note that runTx will actually decode those bytes again. But since
	// this helper is only used in tests/simulation, it's fine.
	bz, err := txEncoder(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}

	gasInfo, result, _, err := app.RunTx(execModeCheck, bz, tx, -1, nil, nil)
	return gasInfo, result, err
}

// Simulate executes a tx in simulate mode to get result and gas info.
func (app *BaseApp) Simulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	gasInfo, result, _, err := app.RunTx(execModeSimulate, txBytes, nil, -1, nil, nil)
	return gasInfo, result, err
}

func (app *BaseApp) SimDeliver(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// See comment for Check().
	bz, err := txEncoder(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}

	gasInfo, result, _, err := app.RunTx(execModeFinalize, bz, tx, -1, nil, nil)
	return gasInfo, result, err
}

func (app *BaseApp) SimTxFinalizeBlock(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// See comment for Check().
	bz, err := txEncoder(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}

	gasInfo, result, _, err := app.RunTx(execModeFinalize, bz, tx, -1, nil, nil)
	return gasInfo, result, err
}

// NewContextLegacy returns a new sdk.Context with the provided header
func (app *BaseApp) NewContextLegacy(isCheckTx bool, header cmtproto.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.stateManager.GetState(execModeCheck).MultiStore, header, true, app.logger).
			WithMinGasPrices(app.gasConfig.MinGasPrices)
	}

	return sdk.NewContext(app.stateManager.GetState(execModeFinalize).MultiStore, header, false, app.logger)
}

// NewContext returns a new sdk.Context with a empty header
func (app *BaseApp) NewContext(isCheckTx bool) sdk.Context {
	return app.NewContextLegacy(isCheckTx, cmtproto.Header{})
}

// NewNextBlockContext sets up the finalize state for the next block and returns
// a context that writes to it. This should be used in tests that need to
// perform state mutations between Commit and the next FinalizeBlock.
func (app *BaseApp) NewNextBlockContext(header cmtproto.Header) sdk.Context {
	app.stateManager.SetState(execModeFinalize, app.cms, header, app.logger, app.streamingManager)
	return app.stateManager.GetState(execModeFinalize).Context()
}

func (app *BaseApp) GetContextForFinalizeBlock(txBytes []byte) sdk.Context {
	return app.getContextForTx(execModeFinalize, txBytes, -1)
}

func (app *BaseApp) GetContextForCheckTx(txBytes []byte) sdk.Context {
	return app.getContextForTx(execModeCheck, txBytes, -1)
}
