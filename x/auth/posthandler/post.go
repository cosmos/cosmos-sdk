package posthandler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HandlerOptions are the options required for constructing a default SDK PostHandler.
type HandlerOptions struct{}

// NewPostHandler returns an empty posthandler chain.
func NewPostHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	postDecorators := []sdk.AnteDecorator{}

	return sdk.ChainAnteDecorators(postDecorators...), nil
}
