package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/gov module sentinel errors
var (
	ErrUnknownProposal       = sdkerrors.Register(ModuleName, 2, "unknown proposal")
	ErrInactiveProposal      = sdkerrors.Register(ModuleName, 3, "inactive proposal")
	ErrAlreadyActiveProposal = sdkerrors.Register(ModuleName, 4, "proposal already active")
	// Errors 5 & 6 are reserved as legacy errors
	// See x/gov/types/v1beta1/errors.go
	ErrInvalidGenesis          = sdkerrors.Register(ModuleName, 8, "invalid genesis state")
	ErrNoProposalHandlerExists = sdkerrors.Register(ModuleName, 9, "no handler exists for proposal type")
	ErrUnroutableProposalMsg   = sdkerrors.Register(ModuleName, 10, "proposal message not recognized by router")
	ErrNoProposalMsgs          = sdkerrors.Register(ModuleName, 11, "no messages proposed")
	ErrInvalidProposalMsg      = sdkerrors.Register(ModuleName, 12, "invalid proposal message")
	ErrInvalidSigner           = sdkerrors.Register(ModuleName, 13, "expected gov account as only signer for proposal message")
	ErrInvalidSignalMsg        = sdkerrors.Register(ModuleName, 14, "signal message is invalid")
	ErrInvalidVote             = sdkerrors.Register(ModuleName, 15, "invalid vote option")
)
