package baseapp

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// SimCheck defines a CheckTx helper function that used in tests and simulations.
func (app *BaseApp) SimCheck(txEncoder sdk.TxEncoder, sdkTx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// CheckTx expects tx bytes as argument, so we encode the tx argument into
	// bytes. Note that CheckTx will actually decode those bytes again. But since
	// this helper is only used in tests/simulation, it's fine.
	bz, err := txEncoder(sdkTx)
	if err != nil {
		return sdk.GasInfo{}, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}

	ctx := app.getContextForTx(runTxModeDeliver, bz)
	res, _, err := app.txHandler.CheckTx(ctx, tx.Request{Tx: sdkTx, TxBytes: bz}, tx.RequestCheckTx{Type: abci.CheckTxType_New})
	gInfo := sdk.GasInfo{GasWanted: uint64(res.GasWanted), GasUsed: uint64(res.GasUsed)}
	if err != nil {
		return gInfo, nil, err
	}

	data, err := makeABCIData(res)
	if err != nil {
		return gInfo, nil, err
	}

	return gInfo, &sdk.Result{Data: data, Log: res.Log, Events: res.Events, MsgResponses: res.MsgResponses}, nil
}

// Simulate executes a tx in simulate mode to get result and gas info.
func (app *BaseApp) Simulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	ctx := app.getContextForTx(runTxModeSimulate, txBytes)
	res, err := app.txHandler.SimulateTx(ctx, tx.Request{TxBytes: txBytes})
	gasInfo := sdk.GasInfo{
		GasWanted: res.GasWanted,
		GasUsed:   res.GasUsed,
	}
	if err != nil {
		return gasInfo, nil, err
	}

	data, err := makeABCIData(res)
	if err != nil {
		return gasInfo, nil, err
	}

	return gasInfo, &sdk.Result{Data: data, Log: res.Log, Events: res.Events, MsgResponses: res.MsgResponses}, nil
}

// SimDeliver defines a DeliverTx helper function that used in tests and
// simulations.
func (app *BaseApp) SimDeliver(txEncoder sdk.TxEncoder, sdkTx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// See comment for Check().
	bz, err := txEncoder(sdkTx)
	if err != nil {
		return sdk.GasInfo{}, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}

	ctx := app.getContextForTx(runTxModeDeliver, bz)
	res, err := app.txHandler.DeliverTx(ctx, tx.Request{Tx: sdkTx, TxBytes: bz})
	gInfo := sdk.GasInfo{GasWanted: uint64(res.GasWanted), GasUsed: uint64(res.GasUsed)}
	if err != nil {
		return gInfo, nil, err
	}

	data, err := makeABCIData(res)
	if err != nil {
		return gInfo, nil, err
	}

	return gInfo, &sdk.Result{Data: data, Log: res.Log, Events: res.Events, MsgResponses: res.MsgResponses}, nil
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
