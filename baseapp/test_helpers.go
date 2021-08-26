package baseapp

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// SimCheck defines a CheckTx helper function that used in tests and simulations.
func (app *BaseApp) SimCheck(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// CheckTx expects tx bytes as argument, so we encode the tx argument into
	// bytes. Note that CheckTx will actually decode those bytes again. But since
	// this helper is only used in tests/simulation, it's fine.
	bz, err := txEncoder(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}

	ctx := app.getContextForTx(runTxModeDeliver, bz)
	res, err := app.txHandler.CheckTx(ctx, tx, abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	gInfo := sdk.GasInfo{GasWanted: uint64(res.GasWanted), GasUsed: uint64(res.GasUsed)}
	if err != nil {
		return gInfo, nil, err
	}

	return gInfo, &sdk.Result{Data: res.Data, Log: res.Log, Events: res.Events}, nil
}

// Simulate executes a tx in simulate mode to get result and gas info.
func (app *BaseApp) Simulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	sdkTx, err := app.txDecoder(txBytes)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	ctx := app.getContextForTx(runTxModeSimulate, txBytes)
	res, err := app.txHandler.SimulateTx(ctx, sdkTx, tx.RequestSimulateTx{TxBytes: txBytes})
	if err != nil {
		return res.GasInfo, nil, err
	}

	return res.GasInfo, res.Result, nil
}

// SimDeliver defines a DeliverTx helper function that used in tests and
// simulations.
func (app *BaseApp) SimDeliver(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// See comment for Check().
	bz, err := txEncoder(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}

	ctx := app.getContextForTx(runTxModeDeliver, bz)
	res, err := app.txHandler.DeliverTx(ctx, tx, abci.RequestDeliverTx{Tx: bz})
	gInfo := sdk.GasInfo{GasWanted: uint64(res.GasWanted), GasUsed: uint64(res.GasUsed)}
	if err != nil {
		return gInfo, nil, err
	}

	return gInfo, &sdk.Result{Data: res.Data, Log: res.Log, Events: res.Events}, nil
}

// Context with current {check, deliver}State of the app used by tests.
func (app *BaseApp) NewContext(isCheckTx bool, header tmproto.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.checkState.ms, header, true, app.logger).
			WithMinGasPrices(app.minGasPrices)
	}

	return sdk.NewContext(app.deliverState.ms, header, false, app.logger)
}

func (app *BaseApp) NewUncachedContext(isCheckTx bool, header tmproto.Header) sdk.Context {
	return sdk.NewContext(app.cms, header, isCheckTx, app.logger)
}
