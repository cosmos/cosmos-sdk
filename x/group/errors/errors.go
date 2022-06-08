package errors

import sdkerrors "cosmossdk.io/errors"

// groupCodespace is the codespace for all errors defined in group package
const groupCodespace = "group"

var (
	ErrEmpty        = sdkerrors.Register(groupCodespace, 2, "value is empty")
	ErrDuplicate    = sdkerrors.Register(groupCodespace, 3, "duplicate value")
	ErrMaxLimit     = sdkerrors.Register(groupCodespace, 4, "limit exceeded")
	ErrType         = sdkerrors.Register(groupCodespace, 5, "invalid type")
	ErrInvalid      = sdkerrors.Register(groupCodespace, 6, "invalid value")
	ErrUnauthorized = sdkerrors.Register(groupCodespace, 7, "unauthorized")
	ErrModified     = sdkerrors.Register(groupCodespace, 8, "modified")
	ErrExpired      = sdkerrors.Register(groupCodespace, 9, "expired")
)
