package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/gov module sentinel errors
var (
	ErrInvalidGranter        = sdkerrors.Register(ModuleName, 1, "invalid granter address")
	ErrInvalidGrantee        = sdkerrors.Register(ModuleName, 2, "invalid grantee address")
	ErrInvalidExpirationTime = sdkerrors.Register(ModuleName, 3, "expiration time of authorization should be more than current time")
)
