package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance module event attributes
const (
	AttributeKeyProposalID         = "proposal_id"
	AttributeValueCategory         = "governance"
	AttributeValueProposalDropped  = "proposal_dropped"  // didn't meet min deposit
	AttributeValueProposalPassed   = "proposal_passed"   // met vote quorum
	AttributeValueProposalRejected = "proposal_rejected" // didn't meet vote quorum
	AttributeValueProposalFailed   = "proposal_failed"   // error on proposal handler
)

// Governance module event types
var (
	EventTypeSubmitProposal  string
	EventTypeProposalDeposit string
	EventTypeProposalVote    string
)

func init() {
	EventTypeSubmitProposal = sdk.GetTypedEventType(&EventSubmitProposal{})
	EventTypeProposalDeposit = sdk.GetTypedEventType(&EventProposalDeposit{})
	EventTypeProposalVote = sdk.GetTypedEventType(&EventVoteProposal{})
}
