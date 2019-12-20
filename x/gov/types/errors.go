package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/gov module errors that reserve codes 500-599
var (
	ErrUnknownProposal         = sdkerrors.Register(ModuleName, 500, "unknown proposal")
	ErrInactiveProposal        = sdkerrors.Register(ModuleName, 501, "inactive proposal")
	ErrAlreadyActiveProposal   = sdkerrors.Register(ModuleName, 502, "proposal already active")
	ErrInvalidProposalContent  = sdkerrors.Register(ModuleName, 503, "invalid proposal content")
	ErrInvalidProposalType     = sdkerrors.Register(ModuleName, 504, "invalid proposal type")
	ErrInvalidVote             = sdkerrors.Register(ModuleName, 505, "invalid vote option")
	ErrInvalidGenesis          = sdkerrors.Register(ModuleName, 506, "invalid genesis state")
	ErrNoProposalHandlerExists = sdkerrors.Register(ModuleName, 507, "no handler exists for proposal type")
)
