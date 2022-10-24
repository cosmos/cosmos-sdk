package types

import (
	sdkerrors "github.com/pointnetwork/cosmos-point-sdk/types/errors"
)

const StoreCodespace = "store"

var ErrInvalidProof = sdkerrors.Register(StoreCodespace, 2, "invalid proof")
