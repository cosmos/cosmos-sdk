package posthandler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// HandlerOptions are the options required for constructing a default SDK PostHandler.
type HandlerOptions struct {
	BankKeeper types.BankKeeper
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewPostHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for posthandler")
	}

	postDecorators := []sdk.AnteDecorator{
		NewTipDecorator(options.BankKeeper),
	}

	return sdk.ChainAnteDecorators(postDecorators...), nil
}
