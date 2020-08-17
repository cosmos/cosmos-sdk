package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC channel sentinel errors
var (
	ErrInvalidPacketTimeout    = sdkerrors.Register(ModuleName, 2, "invalid packet timeout")
	ErrInvalidDenomForTransfer = sdkerrors.Register(ModuleName, 3, "invalid denomination for cross-chain transfer")
	ErrInvalidVersion          = sdkerrors.Register(ModuleName, 4, "invalid ICS20 version")
	ErrInvalidAmount           = sdkerrors.Register(ModuleName, 5, "invalid token amount")
	ErrTraceNotFound           = sdkerrors.Register(ModuleName, 6, "denomination trace not found")
)
