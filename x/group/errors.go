package group

import (
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrEmpty        = errors.Register(ModuleName, 2, "value is empty")
	ErrDuplicate    = errors.Register(ModuleName, 3, "duplicate value")
	ErrMaxLimit     = errors.Register(ModuleName, 4, "limit exceeded")
	ErrType         = errors.Register(ModuleName, 5, "invalid type")
	ErrInvalid      = errors.Register(ModuleName, 6, "invalid value")
	ErrUnauthorized = errors.Register(ModuleName, 7, "unauthorized")
	ErrModified     = errors.Register(ModuleName, 8, "modified")
	ErrExpired      = errors.Register(ModuleName, 9, "expired")
)
