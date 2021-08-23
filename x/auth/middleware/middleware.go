package middleware

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// ComposeMiddlewares compose multiple middlewares on top of a tx.Handler. Last
// middleware is the outermost middleware.
func ComposeMiddlewares(txHandler tx.Handler, middlewares ...tx.Middleware) tx.Handler {
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

// NewDefaultTxHandler defines a TxHandler middleware stacks that should work
// for most applications.
func NewDefaultTxHandler(options TxHandlerOptions) tx.Handler {
	return ComposeMiddlewares(
		NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		newLegacyAnteMiddleware(options.LegacyAnteHandler),
		// Choose which events to index in Tendermint. Make sure no events are
		// emitted outside of this middleware.
		NewIndexEventsTxMiddleware(options.IndexEvents),
		// Recover from panics. Panics outside of this middleware won't be
		// caught, be careful!
		NewRecoveryTxMiddleware(),
		// Set a new GasMeter on sdk.Context.
		//
		// Make sure the Gas middleware is outside of all other middlewares
		// that reads the GasMeter. In our case, the Recovery middleware reads
		// the GasMeter to populate GasInfo.
		NewGasTxMiddleware(),
	)
}
