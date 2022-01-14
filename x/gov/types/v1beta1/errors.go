package v1beta1

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	ErrInvalidProposalContent = sdkerrors.Register(types.ModuleName, 5, "invalid proposal content")
	ErrInvalidProposalType    = sdkerrors.Register(types.ModuleName, 6, "invalid proposal type")
	ErrInvalidVote            = sdkerrors.Register(types.ModuleName, 7, "invalid vote option")
)
