package types

import (
	"cosmossdk.io/errors"
)

const StoreCodespace = "store"

var ErrInvalidProof = sdkerrors.Register(StoreCodespace, 2, "invalid proof")
