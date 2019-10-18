package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authAnte "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/subkeys/internal/keeper"
)

// NewAnteHandler is just like auth.NewAnteHandler, except we use the DeductDelegatedFeeDecorator
// in order to allow payment of fees via a delegation.
func NewAnteHandler(ak authKeeper.AccountKeeper, supplyKeeper authTypes.SupplyKeeper, dk keeper.Keeper, sigGasConsumer authAnte.SignatureVerificationGasConsumer) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		authAnte.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewDelegatedMempoolFeeDecorator(),
		authAnte.NewValidateBasicDecorator(),
		authAnte.NewValidateMemoDecorator(ak),
		authAnte.NewConsumeGasForTxSizeDecorator(ak),
		// DeductDelegatedFeeDecorator will create an empty account if we sign with no tokens but valid validation
		// This must be before SetPubKey, ValidateSigCount, SigVerification, which error if account doesn't exist yet
		NewDeductDelegatedFeeDecorator(ak, supplyKeeper, dk),
		authAnte.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		authAnte.NewValidateSigCountDecorator(ak),
		authAnte.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		authAnte.NewSigVerificationDecorator(ak),
		authAnte.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
	)
}
