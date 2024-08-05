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

	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = errorsmod.Register(RootCodespace, 3, "tx parse error")

	// ErrInsufficientFee is returned if provided fee in tx is less than required fee
	ErrInsufficientFee = errorsmod.Register(RootCodespace, 4, "insufficient fee")

	// ErrAppConfig defines an error occurred if application configuration is misconfigured
	ErrAppConfig = errorsmod.Register(RootCodespace, 5, "error in app.toml")
)
