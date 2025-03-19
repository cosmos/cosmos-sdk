package ante

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper          AccountKeeper
	BankKeeper             types.BankKeeper
	ExtensionOptionChecker ExtensionOptionChecker
	FeegrantKeeper         FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error
	TxFeeChecker           TxFeeChecker
	// UnorderedNonceManager is an opt-in feature for x/auth.
	// When set, this application will be able to receive and process unordered transactions.
	UnorderedNonceManager UnorderedNonceManager
	// UnorderedTxMaxTimeoutDuration is an optional timeout duration for `x/auth`'s unordered transaction feature.
	// When unset, a default of 10 minutes is used.
	UnorderedTxMaxTimeoutDuration time.Duration
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		NewValidateBasicDecorator(),
		NewTxTimeoutHeightDecorator(),
		NewValidateMemoDecorator(options.AccountKeeper),
		NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		NewValidateSigCountDecorator(options.AccountKeeper),
		NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	if options.UnorderedNonceManager != nil {
		var opts []UnorderedTxDecoratorOptions
		if options.UnorderedTxMaxTimeoutDuration != 0 {
			opts = append(opts, WithTimeoutDuration(options.UnorderedTxMaxTimeoutDuration))
		}
		anteDecorators = append(anteDecorators, NewUnorderedTxDecorator(options.UnorderedNonceManager, opts...))
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
