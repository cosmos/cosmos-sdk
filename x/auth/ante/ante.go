package ante

import (
	"cosmossdk.io/core/appmodule"
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth/types"
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	Environment              appmodule.Environment
	AccountKeeper            AccountKeeper
	AccountAbstractionKeeper AccountAbstractionKeeper
	BankKeeper               types.BankKeeper
	ExtensionOptionChecker   ExtensionOptionChecker
	FeegrantKeeper           FeegrantKeeper
	SignModeHandler          *txsigning.HandlerMap
	SigGasConsumer           func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error
	TxFeeChecker             TxFeeChecker
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
		NewSetUpContextDecorator(options.Environment), // outermost AnteDecorator. SetUpContext must be called first
		NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		NewValidateBasicDecorator(options.Environment),
		NewTxTimeoutHeightDecorator(),
		NewValidateMemoDecorator(options.AccountKeeper),
		NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		NewValidateSigCountDecorator(options.AccountKeeper),
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler, options.SigGasConsumer, options.AccountAbstractionKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
