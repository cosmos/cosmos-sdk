package types

import sdkerrors "cosmossdk.io/errors"

const StoreCodespace = "store"

var ErrInvalidProof = sdkerrors.Register(StoreCodespace, 2, "invalid proof")
