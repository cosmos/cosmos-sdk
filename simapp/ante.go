package simapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	feegrantante "github.com/cosmos/cosmos-sdk/x/feegrant/ante"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the
// fee_payer or from fee_granter (if valid grant exist).
func NewAnteHandler(
	ak authkeeper.AccountKeeper, bankKeeper feegranttypes.BankKeeper, feeGrantKeeper feegrantkeeper.Keeper,
	sigGasConsumer authante.SignatureVerificationGasConsumer, signModeHandler signing.SignModeHandler,
) sdk.AnteHandler {

	return sdk.ChainAnteDecorators(
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		authante.NewRejectExtensionOptionsDecorator(),
		authante.NewMempoolFeeDecorator(),
		authante.NewValidateBasicDecorator(),
		authante.TxTimeoutHeightDecorator{},
		authante.NewValidateMemoDecorator(ak),
		authante.NewConsumeGasForTxSizeDecorator(ak),
		// DeductGrantedFeeDecorator will create an empty account if we sign with no
		// tokens but valid validation. This must be before SetPubKey, ValidateSigCount,
		// SigVerification, which error if account doesn't exist yet.
		feegrantante.NewDeductGrantedFeeDecorator(ak, bankKeeper, feeGrantKeeper),
		authante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(ak),
		authante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		authante.NewSigVerificationDecorator(ak, signModeHandler),
		authante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
	)
}
