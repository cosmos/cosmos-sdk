package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/gov module sentinel errors
var (
	ErrUnknownProposal       = errorsmod.Register(ModuleName, 2, "unknown proposal")
	ErrInactiveProposal      = errorsmod.Register(ModuleName, 3, "inactive proposal")
	ErrAlreadyActiveProposal = errorsmod.Register(ModuleName, 4, "proposal already active")
	// Errors 5 & 6 are legacy errors related to v1beta1.Proposal.
	ErrInvalidProposalContent  = errorsmod.Register(ModuleName, 5, "invalid proposal content")
	ErrInvalidProposalType     = errorsmod.Register(ModuleName, 6, "invalid proposal type")
	ErrInvalidVote             = errorsmod.Register(ModuleName, 7, "invalid vote option")
	ErrInvalidGenesis          = errorsmod.Register(ModuleName, 8, "invalid genesis state")
	ErrNoProposalHandlerExists = errorsmod.Register(ModuleName, 9, "no handler exists for proposal type")
	ErrUnroutableProposalMsg   = errorsmod.Register(ModuleName, 10, "proposal message not recognized by router")
	ErrNoProposalMsgs          = errorsmod.Register(ModuleName, 11, "no messages proposed")
	ErrInvalidProposalMsg      = errorsmod.Register(ModuleName, 12, "invalid proposal message")
	ErrInvalidSigner           = errorsmod.Register(ModuleName, 13, "expected gov account as only signer for proposal message")
	ErrInvalidSignalMsg        = errorsmod.Register(ModuleName, 14, "signal message is invalid")
	ErrMetadataTooLong         = errorsmod.Register(ModuleName, 15, "metadata too long")
	ErrMinDepositTooSmall      = errorsmod.Register(ModuleName, 16, "minimum deposit is too small")
	ErrProposalNotFound        = errorsmod.Register(ModuleName, 17, "proposal is not found")
	ErrInvalidProposer         = errorsmod.Register(ModuleName, 18, "invalid proposer")
	ErrNoDeposits              = errorsmod.Register(ModuleName, 19, "no deposits found")
	ErrVotingPeriodEnded       = errorsmod.Register(ModuleName, 20, "voting period already ended")
	ErrInvalidProposal         = errorsmod.Register(ModuleName, 21, "invalid proposal")
)
