package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/crisis module sentinel errors
var (
	ErrNoSender         = errorsmod.Register(ModuleName, 2, "sender address is empty")
	ErrUnknownInvariant = errorsmod.Register(ModuleName, 3, "unknown invariant")
)
