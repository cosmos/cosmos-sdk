package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authnante "github.com/cosmos/cosmos-sdk/x/authn/ante"
	authnkeeper "github.com/cosmos/cosmos-sdk/x/authn/keeper"
	"github.com/cosmos/cosmos-sdk/x/authn/signing"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the
// fee_payer or from fee_granter (if valid grant exist).
func NewAnteHandler(
	ak authnkeeper.AccountKeeper, bankKeeper feegranttypes.BankKeeper, feeGrantKeeper feegrantkeeper.Keeper,
	sigGasConsumer authnante.SignatureVerificationGasConsumer, signModeHandler signing.SignModeHandler,
) sdk.AnteHandler {

	return sdk.ChainAnteDecorators(
		authnante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		authnante.NewRejectExtensionOptionsDecorator(),
		authnante.NewMempoolFeeDecorator(),
		authnante.NewValidateBasicDecorator(),
		authnante.TxTimeoutHeightDecorator{},
		authnante.NewValidateMemoDecorator(ak),
		authnante.NewConsumeGasForTxSizeDecorator(ak),
		NewDeductGrantedFeeDecorator(ak, bankKeeper, feeGrantKeeper),
		authnante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		authnante.NewValidateSigCountDecorator(ak),
		authnante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		authnante.NewSigVerificationDecorator(ak, signModeHandler),
		authnante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
	)
}
