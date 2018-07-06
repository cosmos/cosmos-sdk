package sketchy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
This is just an example to demonstrate a "sketchy" third-party handler module,
to demonstrate the "object capability" model for security.

Since nothing is passed in via arguments to the NewHandler constructor,
it cannot affect the handling of other transaction types.
*/

// Handle all "sketchy" type messages.
type Handler struct{}

// NewHandler returns new Handler
func NewHandler() sdk.Handler { return Handler{} }

// Implements sdk.Handler
func (h Handler) Handle(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	// There's nothing accessible from ctx or msg (even using reflection)
	// that can mutate the state of the application.
	return sdk.Result{}
}

// Implements sdk.Handler
func (h Handler) Type() string {
	return "sketchy"
}
