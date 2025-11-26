package types

import "cosmossdk.io/errors"

var ErrUnauthorized = errors.Register(ModuleName, 1, "unauthorized")
