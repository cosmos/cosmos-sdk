package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/gov module sentinel errors
var (
	ErrUnknownProposal       = sdkerrors.Register(ModuleName, 2, "unknown proposal")
	ErrInactiveProposal      = sdkerrors.Register(ModuleName, 3, "inactive proposal")
	ErrAlreadyActiveProposal = sdkerrors.Register(ModuleName, 4, "proposal already active")

	// Deprecated: these errors are no longer used
	ErrInvalidProposalContent  = sdkerrors.Register(ModuleName, 5, "invalid proposal content")
	ErrInvalidProposalType     = sdkerrors.Register(ModuleName, 6, "invalid proposal type")
	ErrNoProposalHandlerExists = sdkerrors.Register(ModuleName, 9, "no handler exists for proposal type")

	ErrInvalidVote           = sdkerrors.Register(ModuleName, 7, "invalid vote option")
	ErrInvalidGenesis        = sdkerrors.Register(ModuleName, 8, "invalid genesis state")
	ErrUnroutableProposalMsg = sdkerrors.Register(ModuleName, 10, "proposal message has no handler")
	ErrNoProposalMsgs        = sdkerrors.Register(ModuleName, 11, "no messages proposed")
	ErrInvalidSignalMsg      = sdkerrors.Register(ModuleName, 12, "invalid signal message")
)
