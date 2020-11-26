package simapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantante "github.com/cosmos/cosmos-sdk/x/feegrant/ante"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(
	ak authkeeper.AccountKeeper, supplyKeeper feegrant.SupplyKeeper, feeGrantKeeper feegrant.Keeper,
	sigGasConsumer authante.SignatureVerificationGasConsumer,
) sdk.AnteHandler {

	txConfig := legacytx.StdTxConfig{Cdc: codec.NewLegacyAmino()}
	return sdk.ChainAnteDecorators(
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		authante.NewMempoolFeeDecorator(),
		authante.NewValidateBasicDecorator(),
		authante.NewValidateMemoDecorator(ak),
		authante.NewConsumeGasForTxSizeDecorator(ak),
		// DeductGrantedFeeDecorator will create an empty account if we sign with no
		// tokens but valid validation. This must be before SetPubKey, ValidateSigCount,
		// SigVerification, which error if account doesn't exist yet.
		feegrantante.NewDeductGrantedFeeDecorator(ak, supplyKeeper, feeGrantKeeper),
		authante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(ak),
		authante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		authante.NewSigVerificationDecorator(ak, txConfig.SignModeHandler()),
		authante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
	)
}
