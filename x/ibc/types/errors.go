package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ibc module common sentinel errors
var (
	ErrInvalidHeight  = sdkerrors.Register(ModuleName, 2, "invalid height")
	ErrInvalidVersion = sdkerrors.Register(ModuleName, 3, "invalid version")
)
