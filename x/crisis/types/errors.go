package types

import "github.com/cosmos/cosmos-sdk/errors/v3"

// x/crisis module sentinel errors
var (
	ErrNoSender         = errors.Register(ModuleName, 2, "sender address is empty")
	ErrUnknownInvariant = errors.Register(ModuleName, 3, "unknown invariant")
)
