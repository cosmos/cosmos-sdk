package types

import (
	"errors"

	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrInvalidAddress = sdkerrors.Register(ModuleName, 2, "invalid account address")
	ErrNonExistOwner  = sdkerrors.Register(ModuleName, 3, "owner account does not exist")
	errInvalidMessage = errors.New("invalid message")
)

// IsRoutingError returns true if the error is a routing error,
// which typically occurs when a message cannot be matched to a handler.
func IsRoutingError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, errInvalidMessage)
}
