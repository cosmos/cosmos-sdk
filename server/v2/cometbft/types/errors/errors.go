package errors

import (
	errorsmod "cosmossdk.io/errors"
)

// RootCodespace is the codespace for all errors defined in this package
const RootCodespace = "cometbft-sdk"

var (
	// ErrUnknownRequest to doc
	ErrUnknownRequest = errorsmod.Register(RootCodespace, 1, "unknown request")

	// ErrInvalidRequest defines an ABCI typed error where the request contains
	// invalid data.
	ErrInvalidRequest = errorsmod.Register(RootCodespace, 2, "invalid request")
)
