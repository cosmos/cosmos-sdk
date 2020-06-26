package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC local-host client sentinel errors
var (
	ErrInvalidInitialization = sdkerrors.Register(SubModuleName, 2, "must initialize local host with context")
	ErrInvalidUpdate         = sdkerrors.Register(SubModuleName, 3, "must update local host on begin block with context")
	ErrInvalidMisbehaviour   = sdkerrors.Register(SubModuleName, 4, "local-host does not accept misbehaviour")
)
