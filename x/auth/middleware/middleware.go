package middleware

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// Enum mode for CheckTx and DeliverTx
type runTxMode uint8

const (
	runTxModeCheck    runTxMode = iota // Check a transaction
	runTxModeReCheck                   // Recheck a (pending) transaction after a commit
	runTxModeSimulate                  // Simulate a transaction
	runTxModeDeliver                   // Deliver a transaction
)

// ComposeTxMiddleware compose multiple middlewares on top of a TxHandler. First
// middleware is the outermost middleware (i.e. gets run first).
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
