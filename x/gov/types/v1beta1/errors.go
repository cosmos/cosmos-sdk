package v1beta1

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidProposalContent = sdkerrors.Register(ModuleName, 5, "invalid proposal content")
	ErrInvalidProposalType = sdkerrors.Register(ModuleName, 6, "invalid proposal type")
	ErrInvalidVote         = sdkerrors.Register(ModuleName, 7, "invalid vote option")
)
