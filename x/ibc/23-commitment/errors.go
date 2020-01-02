package commitment

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SubModuleName is the error codespace
const SubModuleName string = "ibc/commitment"

// IBC connection sentinel errors
var (
	ErrInvalidProof  = sdkerrors.Register(SubModuleName, 1, "invalid proof")
	ErrInvalidPrefix = sdkerrors.Register(SubModuleName, 2, "invalid prefix")
)
