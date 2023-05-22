package errors

import (
	errorsmod "cosmossdk.io/errors"
)

// groupCodespace is the codespace for all errors defined in group package
const groupCodespace = "group"

var (
	ErrEmpty        = errorsmod.Register(groupCodespace, 2, "value is empty")
	ErrDuplicate    = errorsmod.Register(groupCodespace, 3, "duplicate value")
	ErrMaxLimit     = errorsmod.Register(groupCodespace, 4, "limit exceeded")
	ErrType         = errorsmod.Register(groupCodespace, 5, "invalid type")
	ErrInvalid      = errorsmod.Register(groupCodespace, 6, "invalid value")
	ErrUnauthorized = errorsmod.Register(groupCodespace, 7, "unauthorized")
	ErrModified     = errorsmod.Register(groupCodespace, 8, "modified")
	ErrExpired      = errorsmod.Register(groupCodespace, 9, "expired")
)
