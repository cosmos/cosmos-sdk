// nolint
package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ActionProposalDropped  = []byte("proposal-dropped")
	ActionProposalPassed   = []byte("proposal-passed")
	ActionProposalRejected = []byte("proposal-rejected")

	Action            = sdk.TagAction
	Proposer          = "proposer"
	ProposalID        = "proposal-id"
	VotingPeriodStart = "voting-period-start"
	Depositor         = "depositor"
	Voter             = "voter"
)
