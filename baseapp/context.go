package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

// NewContext returns a new Context suitable for AnteHandler (and indirectly Handler) processing.
// NOTE: txBytes may be nil to support TestApp.RunCheckTx
// and TestApp.RunDeliverTx.
func (app *BaseApp) NewContext(isCheckTx bool, txBytes []byte) sdk.Context {
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

	return sdk.NewContext(store, *app.header, isCheckTx, txBytes)
}

// context used during genesis
func (app *BaseApp) GenesisContext(txBytes []byte) sdk.Context {
	store := app.msDeliver
	if store == nil {
		panic("BaseApp.NewContext() requires BeginBlock(): missing store")
	}
	return sdk.NewContext(store, abci.Header{}, false, txBytes)
}
