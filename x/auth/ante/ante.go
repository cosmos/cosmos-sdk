package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// HandlerOptions are the options for ante handler build
type HandlerOptions struct {
	FeegrantKeeper FeegrantKeeper
	SigGasConsumer func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(
	ak AccountKeeper, bk types.BankKeeper,
	signModeHandler authsigning.SignModeHandler,
	anteHandlerOptions HandlerOptions,
) sdk.AnteHandler {
	var sigGasConsumer = anteHandlerOptions.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewRejectExtensionOptionsDecorator(),
		NewMempoolFeeDecorator(),
		NewValidateBasicDecorator(),
		TxTimeoutHeightDecorator{},
		NewValidateMemoDecorator(ak),
		NewConsumeGasForTxSizeDecorator(ak),
		NewDeductFeeDecorator(ak, bk, anteHandlerOptions.FeegrantKeeper),
		NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		NewValidateSigCountDecorator(ak),
		NewSigGasConsumeDecorator(ak, sigGasConsumer),
		NewSigVerificationDecorator(ak, signModeHandler),
		NewIncrementSequenceDecorator(ak),
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}
