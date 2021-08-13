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

type DefaultTxHandlerOptions struct {
	Debug bool

	LegacyRouter     sdk.Router
	MsgServiceRouter *MsgServiceRouter

	LegacyAnteHandler sdk.AnteHandler
}

func NewDefaultTxHandler(options DefaultTxHandlerOptions) tx.TxHandler {
	return ComposeTxMiddleware(
		NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		newLegacyAnteMiddleware(options.LegacyAnteHandler),
		NewPanicTxMiddleware(),
		NewErrorTxMiddleware(options.Debug),
	)
}
