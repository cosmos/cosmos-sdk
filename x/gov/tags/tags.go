package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance tags
var (
	ActionProposalDropped   = "proposal-dropped"
	ActionProposalPassed    = "proposal-passed"
	ActionProposalRejected  = "proposal-rejected"
	ActionProposalSubmitted = "proposal-submitted"
	ActionProposalVote      = "proposal-vote"
	ActionProposalDeposit   = "proposal-deposit"

	Action            = sdk.TagAction
	Proposer          = "proposer"
	ProposalID        = "proposal-id"
	VotingPeriodStart = "voting-period-start"
	Depositor         = "depositor"
	Voter             = "voter"
	ProposalResult    = "proposal-result"
)
