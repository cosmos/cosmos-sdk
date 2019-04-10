package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance tags
const (
	ActionProposalDropped  = "proposal-dropped"
	ActionProposalPassed   = "proposal-passed"
	ActionProposalRejected = "proposal-rejected"
	ActionProposalFailed   = "proposal-failed"
	TxCategory             = "governance"

	Action            = sdk.TagAction
	Category          = sdk.TagCategory
	Proposer          = "proposer"
	ProposalID        = "proposal-id"
	VotingPeriodStart = "voting-period-start"
	Depositor         = "depositor"
	Voter             = "voter"
	ProposalResult    = "proposal-result"
)
