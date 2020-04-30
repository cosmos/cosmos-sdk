package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	SubModuleName = "solo machine"
)

var (
	ErrInvalidHeader = sdkerrors.Register(SubModuleName, 1, "invalid header")
)
