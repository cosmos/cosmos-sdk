package types

import "cosmossdk.io/errors"

var ErrInvalidSigner = errors.Register(ModuleName, 2, "expected authority account as only signer for community pool spend message")
