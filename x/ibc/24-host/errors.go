package host

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBCCodeSpace is the codespace for all errors defined in the ibc module
const IBCCodeSpace = "ibc"

var (
	// ErrInvalidID is returned if identifier string is invalid
	ErrInvalidID = sdkerrors.Register(IBCCodeSpace, 1, "invalid identifier")

	// ErrInvalidPath is returned if path string is invalid
	ErrInvalidPath = sdkerrors.Register(IBCCodeSpace, 2, "invalid path")

	// ErrInvalidPacket is returned if packets embedded in msg are invalid
	ErrInvalidPacket = sdkerrors.Register(IBCCodeSpace, 3, "invalid packet extracted from msg")
)
