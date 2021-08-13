package baseapp

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (app *BaseApp) Check(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// CheckTx expects tx bytes as argument, so we encode the tx argument into
	// bytes. Note that CheckTx will actually decode those bytes again. But since
	// this helper is only used in tests/simulation, it's fine.
	bz, err := txEncoder(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}
	res := app.CheckTx(abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	return sdk.GasInfo{GasWanted: uint64(res.GasWanted), GasUsed: uint64(res.GasUsed)},
		&sdk.Result{Data: res.Data, Log: res.Log, Events: res.Events},
		nil
}

func (app *BaseApp) Simulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: bz})
	return sdk.GasInfo{GasWanted: uint64(res.GasWanted), GasUsed: uint64(res.GasUsed)},
		&sdk.Result{Data: res.Data, Log: res.Log, Events: res.Events},
		nil
}

func (app *BaseApp) Deliver(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	// See comment for Check().
	bz, err := txEncoder(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err)
	}
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: bz})
	return sdk.GasInfo{GasWanted: uint64(res.GasWanted), GasUsed: uint64(res.GasUsed)},
		&sdk.Result{Data: res.Data, Log: res.Log, Events: res.Events},
		nil
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
