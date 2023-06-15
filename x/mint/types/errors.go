package types

import "cosmossdk.io/errors"

var ErrInvalidSigner = errors.Register(ModuleName, 1, "invalid signer")
