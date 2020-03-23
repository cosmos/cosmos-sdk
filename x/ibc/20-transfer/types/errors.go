package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC channel sentinel errors
var (
	ErrInvalidPacketTimeout = sdkerrors.Register(ModuleName, 1, "invalid packet timeout")
)
