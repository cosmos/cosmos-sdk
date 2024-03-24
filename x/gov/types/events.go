package types

// Governance module event types
const (
	EventTypeSubmitProposal   = "submit_proposal"
	EventTypeProposalDeposit  = "proposal_deposit"
	EventTypeProposalVote     = "proposal_vote"
	EventTypeInactiveProposal = "inactive_proposal"
	EventTypeActiveProposal   = "active_proposal"
	EventTypeCancelProposal   = "cancel_proposal"

<<<<<<< HEAD
	AttributeKeyProposalResult              = "proposal_result"
	AttributeKeyVoter                       = "voter"
	AttributeKeyOption                      = "option"
	AttributeKeyProposalID                  = "proposal_id"
	AttributeKeyProposalMessages            = "proposal_messages" // Msg type_urls in the proposal
	AttributeKeyVotingPeriodStart           = "voting_period_start"
	AttributeKeyProposalLog                 = "proposal_log"                // log of proposal execution
	AttributeValueProposalDropped           = "proposal_dropped"            // didn't meet min deposit
	AttributeValueProposalPassed            = "proposal_passed"             // met vote quorum
	AttributeValueProposalRejected          = "proposal_rejected"           // didn't meet vote quorum
	AttributeValueExpeditedProposalRejected = "expedited_proposal_rejected" // didn't meet expedited vote quorum
	AttributeValueProposalFailed            = "proposal_failed"             // error on proposal handler
	AttributeValueProposalCanceled          = "proposal_canceled"           // error on proposal handler
=======
	AttributeKeyProposalResult       = "proposal_result"
	AttributeKeyVoter                = "voter"
	AttributeKeyOption               = "option"
	AttributeKeyProposalID           = "proposal_id"
	AttributeKeyProposalMessages     = "proposal_messages" // Msg type_urls in the proposal
	AttributeKeyVotingPeriodStart    = "voting_period_start"
	AttributeKeyProposalLog          = "proposal_log"           // log of proposal execution
	AttributeKeyProposalDepositError = "proposal_deposit_error" // error on proposal deposit refund/burn
	AttributeKeyProposalProposer     = "proposal_proposer"      // account address of the proposer

	AttributeValueProposalDropped            = "proposal_dropped"             // didn't meet min deposit
	AttributeValueProposalPassed             = "proposal_passed"              // met vote quorum
	AttributeValueProposalRejected           = "proposal_rejected"            // didn't meet vote quorum
	AttributeValueExpeditedProposalRejected  = "expedited_proposal_rejected"  // didn't meet expedited vote quorum
	AttributeValueOptimisticProposalRejected = "optimistic_proposal_rejected" // didn't meet optimistic vote quorum
	AttributeValueProposalFailed             = "proposal_failed"              // error on proposal handler
	AttributeValueProposalCanceled           = "proposal_canceled"            // error on proposal handler
>>>>>>> 45994391a (feat(x/gov): emit proposer address in submit proposal event (#19842))

	AttributeKeyProposalType   = "proposal_type"
	AttributeSignalTitle       = "signal_title"
	AttributeSignalDescription = "signal_description"
)
