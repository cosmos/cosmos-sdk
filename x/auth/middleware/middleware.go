package middleware

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// ComposeMiddlewares compose multiple middlewares on top of a tx.Handler. The
// middleware order in the variadic arguments is from inside to outside.
//
// Example: Given a base tx.Handler H, and two middlewares A and B, the
// middleware stack:
// ```
// A.pre
//   B.pre
//     H
//   B.post
// A.post
// ```
// is created by calling `ComposeMiddlewares(H, A, B)`.
func ComposeMiddlewares(txHandler tx.Handler, middlewares ...tx.Middleware) tx.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		txHandler = middlewares[i](txHandler)
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
		// Set a new GasMeter on sdk.Context.
		//
		// Make sure the Gas middleware is outside of all other middlewares
		// that reads the GasMeter. In our case, the Recovery middleware reads
		// the GasMeter to populate GasInfo.
		GasTxMiddleware,
		// Recover from panics. Panics outside of this middleware won't be
		// caught, be careful!
		RecoveryTxMiddleware,
		// Choose which events to index in Tendermint. Make sure no events are
		// emitted outside of this middleware.
		NewIndexEventsTxMiddleware(options.IndexEvents),
		// Reject all extension options which can optionally be included in the
		// tx.
		RejectExtensionOptionsMiddleware,
		MempoolFeeMiddleware,
		// Temporary middleware to bundle antehandlers.
		// TODO Remove in https://github.com/cosmos/cosmos-sdk/issues/9585.
		newLegacyAnteMiddleware(options.LegacyAnteHandler),
	)
}
