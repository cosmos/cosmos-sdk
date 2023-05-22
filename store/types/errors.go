package types

import (
	errorsmod "cosmossdk.io/errors"
)

const StoreCodespace = "store"

var ErrInvalidProof = errorsmod.Register(StoreCodespace, 2, "invalid proof")
