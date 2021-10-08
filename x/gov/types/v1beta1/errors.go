package v1beta1

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// Deprecated
	ErrInvalidProposalContent = sdkerrors.Register(ModuleName, 5, "invalid proposal content")
	// Deprecated
	ErrInvalidProposalType = sdkerrors.Register(ModuleName, 6, "invalid proposal type")
)
