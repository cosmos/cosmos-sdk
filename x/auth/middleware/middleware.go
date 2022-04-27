package middleware

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// ComposeMiddlewares compose multiple middlewares on top of a tx.Handler. The
// middleware order in the variadic arguments is from outer to inner.
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

	// TxDecoder is used to decode the raw tx bytes into a sdk.Tx.
	TxDecoder sdk.TxDecoder

	// IndexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs Tendermint what to index. If empty, all events will be indexed.
	IndexEvents map[string]struct{}

	LegacyRouter     sdk.Router
	MsgServiceRouter *MsgServiceRouter

	AccountKeeper          AccountKeeper
	BankKeeper             types.BankKeeper
	FeegrantKeeper         FeegrantKeeper
	SignModeHandler        authsigning.SignModeHandler
	SigGasConsumer         func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error
	ExtensionOptionChecker ExtensionOptionChecker
	TxFeeChecker           TxFeeChecker
}

// NewDefaultTxHandler defines a TxHandler middleware stacks that should work
// for most applications.
func NewDefaultTxHandler(options TxHandlerOptions) (tx.Handler, error) {
	if options.TxDecoder == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "txDecoder is required for middlewares")
	}

	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for middlewares")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for middlewares")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for middlewares")
	}

	return ComposeMiddlewares(
		NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		NewTxDecoderMiddleware(options.TxDecoder),
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
		// Reject all extension options other than the ones needed by the feemarket.
		NewExtensionOptionsMiddleware(options.ExtensionOptionChecker),
		ValidateBasicMiddleware,
		TxTimeoutHeightMiddleware,
		ValidateMemoMiddleware(options.AccountKeeper),
		ConsumeTxSizeGasMiddleware(options.AccountKeeper),
		// No gas should be consumed in any middleware above in a "post" handler part. See
		// ComposeMiddlewares godoc for details.
		// `DeductFeeMiddleware` and `IncrementSequenceMiddleware` should be put outside of `WithBranchedStore` middleware,
		// so their storage writes are not discarded when tx fails.
		DeductFeeMiddleware(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		SetPubKeyMiddleware(options.AccountKeeper),
		ValidateSigCountMiddleware(options.AccountKeeper),
		SigGasConsumeMiddleware(options.AccountKeeper, options.SigGasConsumer),
		SigVerificationMiddleware(options.AccountKeeper, options.SignModeHandler),
		IncrementSequenceMiddleware(options.AccountKeeper),
		// Creates a new MultiStore branch, discards downstream writes if the downstream returns error.
		// These kinds of middlewares should be put under this:
		// - Could return error after messages executed succesfully.
		// - Storage writes should be discarded together when tx failed.
		WithBranchedStore,
		// Consume block gas. All middlewares whose gas consumption after their `next` handler
		// should be accounted for, should go below this middleware.
		ConsumeBlockGasMiddleware,
		NewTipMiddleware(options.BankKeeper),
	), nil
}
