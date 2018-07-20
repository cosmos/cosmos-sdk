// nolint
package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ActionSubmitProposal = []byte("submit-proposal")

	Action = sdk.TagAction
	//SrcValidator = sdk.TagSrcValidator
	//DstValidator = sdk.TagDstValidator
	//Delegator    = sdk.TagDelegator
	Proposer          = "proposer"
	ProposalID        = "proposal-ID"
	VotingPeriodStart = "VotingPeriodStart"
)
