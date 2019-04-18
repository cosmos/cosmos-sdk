package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance tags
const (
	ActionProposalDropped  = "proposal-dropped"  // didn't meet min deposit
	ActionProposalPassed   = "proposal-passed"   // meet vote quorum
	ActionProposalRejected = "proposal-rejected" // didn't meet vote quorum
	ActionProposalFailed   = "proposal-failed"   // error on proposal handler
	TxCategory             = "governance"

	ProposalID        = "proposal-id"
	VotingPeriodStart = "voting-period-start"
	ProposalResult    = "proposal-result"
)

// SDK tag aliases
var (
	Action   = sdk.TagAction
	Category = sdk.TagCategory
	Sender   = sdk.TagSender
)
