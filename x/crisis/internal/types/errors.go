package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/crisis module errors that reserve codes 200-299
var (
	ErrNoSender         = sdkerrors.Register(ModuleName, 200, "sender address is empty")
	ErrUnknownInvariant = sdkerrors.Register(ModuleName, 201, "unknown invariant")
)
