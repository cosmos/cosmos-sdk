package types

import "cosmossdk.io/errors"

// x/crisis module sentinel errors
var (
	ErrNoSender         = errors.Register(ModuleName, 2, "sender address is empty")
	ErrUnknownInvariant = errors.Register(ModuleName, 3, "unknown invariant")
	ErrInvalidSigner    = errors.Register(ModuleName, 4, "expected authority account as only signer for proposal message")
)
