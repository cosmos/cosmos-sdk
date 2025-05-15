package types

import "github.com/cosmos/cosmos-sdk/errors/v3"

var ErrInvalidSigner = errors.Register(ModuleName, 2, "expected authority account as only signer for community pool spend message")
