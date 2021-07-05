package group

import (
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrEmpty        = errors.Register(ModuleName, 202, "value is empty")
	ErrDuplicate    = errors.Register(ModuleName, 203, "duplicate value")
	ErrMaxLimit     = errors.Register(ModuleName, 204, "limit exceeded")
	ErrType         = errors.Register(ModuleName, 205, "invalid type")
	ErrInvalid      = errors.Register(ModuleName, 206, "invalid value")
	ErrUnauthorized = errors.Register(ModuleName, 207, "unauthorized")
	ErrModified     = errors.Register(ModuleName, 208, "modified")
	ErrExpired      = errors.Register(ModuleName, 209, "expired")
)
