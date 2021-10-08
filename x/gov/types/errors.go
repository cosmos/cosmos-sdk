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
	ErrInvalidVote             = sdkerrors.Register(ModuleName, 7, "invalid vote option")
	ErrInvalidGenesis          = sdkerrors.Register(ModuleName, 8, "invalid genesis state")
	ErrNoProposalHandlerExists = sdkerrors.Register(ModuleName, 9, "no handler exists for proposal type")
	ErrUnroutableProposalMsg   = sdkerrors.Register(ModuleName, 10, "proposal message has no handler")
	ErrNoProposalMsgs          = sdkerrors.Register(ModuleName, 11, "no messages proposed")
	ErrInvalidSigner           = sdkerrors.Register(ModuleName, 12, "expected gov account as only signer for proposal message")
	ErrInvalidSignalMsg        = sdkerrors.Register(ModuleName, 13, "signal message is invalid")
)
