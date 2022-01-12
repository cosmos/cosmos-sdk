package v1beta1

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidProposalContent = sdkerrors.Register(moduleName, 5, "invalid proposal content")
	ErrInvalidProposalType    = sdkerrors.Register(moduleName, 6, "invalid proposal type")
	ErrInvalidVote            = sdkerrors.Register(moduleName, 7, "invalid vote option")
)
