package simapp

import (
	"errors"

	circuitante "cosmossdk.io/x/circuit/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/ante/unorderedtx"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	ante.HandlerOptions
	CircuitKeeper circuitante.CircuitBreaker
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errors.New("account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errors.New("bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errors.New("sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(options.Environment, options.ConsensusKeeper), // outermost AnteDecorator. SetUpContext must be called first
		circuitante.NewCircuitBreakerDecorator(options.CircuitKeeper),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(options.Environment),
		ante.NewTxTimeoutHeightDecorator(options.Environment),
		ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, options.UnorderedTxManager, options.Environment, ante.DefaultSha256Cost),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler, options.SigGasConsumer, options.AccountAbstractionKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
