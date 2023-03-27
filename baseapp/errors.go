package baseapp

import errorsmod "cosmossdk.io/errors"

// RootCodespace is the codespace for all errors defined in this package
const RootCodespace = "baseapp"

var (
	// ErrMultiMsgFailure defines an error for when a multi-message transaction fails.
	ErrMultiMsgFailure = errorsmod.Register(RootCodespace, 2, "all messages failed")

	// ErrBadNonAtomicSignature defines an error for when a non-atomic transaction has a legacy or invalid signature
	ErrBadNonAtomicSignature = errorsmod.Register(RootCodespace, 3, "bad signature for non-atomic txs")
)
