package tags

// Governance tags
const (
	ActionProposalDropped  = "proposal_dropped"  // didn't meet min deposit
	ActionProposalPassed   = "proposal_passed"   // meet vote quorum
	ActionProposalRejected = "proposal_rejected" // didn't meet vote quorum
	ActionProposalFailed   = "proposal_failed"   // error on proposal handler
	TxCategory             = "governance"
	Deposit                = "deposit"
	Option                 = "option"

	ProposalID        = "proposal_id"
	VotingPeriodStart = "voting_period_start"
	ProposalResult    = "proposal_result"
	ProposalDeposit   = "proposal_deposit"
	ProposalVote      = "proposal_vote"
	SubmitProposal    = "submit_proposal"
	Amount            = "amount"

	InactiveProposal = "inactive_proposal"
	ActiveProposal   = "active_proposal"
)
