package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(ak keeper.AccountKeeper, supplyKeeper types.SupplyKeeper, sigGasConsumer SignatureVerificationGasConsumer) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		NewSetupDecorator(),
		NewMempoolFeeDecorator(),
		NewValidateBasicDecorator(),
		NewValidateMemoDecorator(ak),
		NewConsumeGasForTxSizeDecorator(ak),
		NewDeductFeeDecorator(ak, supplyKeeper),
		NewSetPubKeyDecorator(ak),
		NewValidateSigCountDecorator(ak),
		NewSigGasConsumeDecorator(ak, sigGasConsumer),
		NewSigVerificationDecorator(ak),
	)
}
