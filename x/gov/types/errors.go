package types

import (
	"cosmossdk.io/errors"
)

// x/gov module sentinel errors
var (
	ErrInactiveProposal      = errors.Register(ModuleName, 3, "inactive proposal")
	ErrAlreadyActiveProposal = errors.Register(ModuleName, 4, "proposal already active")
	// Errors 5 & 6 are legacy errors related to v1beta1.Proposal.
	ErrInvalidProposalContent         = errors.Register(ModuleName, 5, "invalid proposal content")
	ErrInvalidProposalType            = errors.Register(ModuleName, 6, "invalid proposal type")
	ErrInvalidVote                    = errors.Register(ModuleName, 7, "invalid vote option")
	ErrInvalidGenesis                 = errors.Register(ModuleName, 8, "invalid genesis state")
	ErrNoProposalHandlerExists        = errors.Register(ModuleName, 9, "no handler exists for proposal type")
	ErrUnroutableProposalMsg          = errors.Register(ModuleName, 10, "proposal message not recognized by router")
	ErrNoProposalMsgs                 = errors.Register(ModuleName, 11, "no messages proposed")
	ErrInvalidProposalMsg             = errors.Register(ModuleName, 12, "invalid proposal message")
	ErrInvalidSigner                  = errors.Register(ModuleName, 13, "expected gov account as only signer for proposal message")
	ErrMetadataTooLong                = errors.Register(ModuleName, 15, "metadata too long")
	ErrMinDepositTooSmall             = errors.Register(ModuleName, 16, "minimum deposit is too small")
	ErrInvalidProposer                = errors.Register(ModuleName, 18, "invalid proposer")
	ErrVotingPeriodEnded              = errors.Register(ModuleName, 20, "voting period already ended")
	ErrInvalidProposal                = errors.Register(ModuleName, 21, "invalid proposal")
	ErrSummaryTooLong                 = errors.Register(ModuleName, 22, "summary too long")
	ErrInvalidDepositDenom            = errors.Register(ModuleName, 23, "invalid deposit denom")
	ErrInvalidConstitutionAmendment   = errors.Register(ModuleName, 170, "invalid constitution amendment")
	ErrGovernorExists                 = errors.Register(ModuleName, 300, "governor already exists")
	ErrGovernorNotFound               = errors.Register(ModuleName, 310, "governor not found")
	ErrInvalidGovernorStatus          = errors.Register(ModuleName, 320, "invalid governor status")
	ErrGovernanceDelegationExists     = errors.Register(ModuleName, 330, "governance delegation already exists")
	ErrGovernanceDelegationNotFound   = errors.Register(ModuleName, 340, "governance delegation not found")
	ErrInvalidGovernanceDescription   = errors.Register(ModuleName, 350, "invalid governance description")
	ErrDelegatorIsGovernor            = errors.Register(ModuleName, 360, "cannot explicitly manage governance delegations, delegator is an active governor")
	ErrGovernorStatusEqual            = errors.Register(ModuleName, 370, "cannot change governor status to the same status")
	ErrGovernorStatusChangePeriod     = errors.Register(ModuleName, 380, "governor status change period not elapsed")
	ErrInsufficientGovernorDelegation = errors.Register(ModuleName, 390, "insufficient governor self-delegation")
)
