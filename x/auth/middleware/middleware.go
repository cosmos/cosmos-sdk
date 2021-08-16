package middleware

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// ComposeTxMiddleware compose multiple middlewares on top of a TxHandler. Last
// middleware is the outermost middleware.
func ComposeTxMiddleware(txHandler tx.TxHandler, middlewares ...tx.TxMiddleware) tx.TxHandler {
	for _, m := range middlewares {
		txHandler = m(txHandler)
	}

	return txHandler
}

type TxHandlerOptions struct {
	Debug bool
	// IndexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs Tendermint what to index. If empty, all events will be indexed.
	IndexEvents map[string]struct{}

	LegacyRouter     sdk.Router
	MsgServiceRouter *MsgServiceRouter

	LegacyAnteHandler sdk.AnteHandler
}

func NewDefaultTxHandler(options TxHandlerOptions) tx.TxHandler {
	return ComposeTxMiddleware(
		NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		newLegacyAnteMiddleware(options.LegacyAnteHandler),
		// Make sure no events are emitted outside of this middleware.
		NewIndexEventsTxMiddleware(options.IndexEvents),
		NewPanicTxMiddleware(),
		NewErrorTxMiddleware(options.Debug),
	)
}
