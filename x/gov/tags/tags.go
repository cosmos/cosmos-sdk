package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance tags
var (
	ActionProposalDropped   = []byte("proposal-dropped")
	ActionProposalPassed    = []byte("proposal-passed")
	ActionProposalRejected  = []byte("proposal-rejected")
	ActionProposalSubmitted = []byte("proposal-submitted")
	ActionProposalVote      = []byte("proposal-vote")
	ActionProposalDeposit   = []byte("proposal-deposit")

	Action            = sdk.TagAction
	Proposer          = "proposer"
	ProposalID        = "proposal-id"
	VotingPeriodStart = "voting-period-start"
	Depositor         = "depositor"
	Voter             = "voter"
)
