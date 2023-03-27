package baseapp

import errorsmod "cosmossdk.io/errors"

// RootCodespace is the codespace for all errors defined in this package
const RootCodespace = "baseapp"

var (
	// ErrMultiMsgFailure defines an error for when a multi-message transaction fails.
	ErrMultiMsgFailure = errorsmod.Register(RootCodespace, 2, "all messages failed")
)
