package middleware

import "github.com/cosmos/cosmos-sdk/types/tx"

// Enum mode for CheckTx and DeliverTx
type runTxMode uint8

const (
	runTxModeCheck    runTxMode = iota // Check a transaction
	runTxModeReCheck                   // Recheck a (pending) transaction after a commit
	runTxModeSimulate                  // Simulate a transaction
	runTxModeDeliver                   // Deliver a transaction
)

// ComposeTxMiddleware compose multiple middlewares on top of a TxHandler. Last
// middleware is the outermost middleware.
func ComposeTxMiddleware(txHandler tx.TxHandler, middlewares ...tx.TxMiddleware) tx.TxHandler {
	for _, m := range middlewares {
		txHandler = m(txHandler)
	}

	return txHandler
}

func NewDefaultTxHandler(debug bool) tx.TxHandler {
	return ComposeTxMiddleware(
		NewRunMsgsTxHandler(),
		// add antehandlers here
		validateMiddleware,
		NewPanicTxMiddleware(debug),
	)
}
