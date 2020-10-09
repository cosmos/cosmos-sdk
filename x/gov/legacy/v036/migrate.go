package v036

import (
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. This migration flattens the deposits and votes and updates the
// proposal content to the new
func Migrate(oldGenState v034gov.GenesisState) GenesisState {
	deposits := make(v034gov.Deposits, len(oldGenState.Deposits))
	for i, deposit := range oldGenState.Deposits {
		deposits[i] = deposit.Deposit
	}

	votes := make(v034gov.Votes, len(oldGenState.Votes))
	for i, vote := range oldGenState.Votes {
		votes[i] = vote.Vote
	}

	proposals := make([]Proposal, len(oldGenState.Proposals))
	for i, proposal := range oldGenState.Proposals {
		proposals[i] = Proposal{
			Content:          migrateContent(proposal.ProposalContent),
			ProposalID:       proposal.ProposalID,
			Status:           proposal.Status,
			FinalTallyResult: proposal.FinalTallyResult,
			SubmitTime:       proposal.SubmitTime,
			DepositEndTime:   proposal.DepositEndTime,
			TotalDeposit:     proposal.TotalDeposit,
			VotingStartTime:  proposal.VotingStartTime,
			VotingEndTime:    proposal.VotingEndTime,
		}
	}

	return NewGenesisState(
		oldGenState.StartingProposalID, deposits, votes, proposals,
		oldGenState.DepositParams, oldGenState.VotingParams, oldGenState.TallyParams,
	)
}

func migrateContent(proposalContent v034gov.ProposalContent) (content Content) {
	switch proposalContent.ProposalType() {
	case v034gov.ProposalTypeText:
		return NewTextProposal(proposalContent.GetTitle(), proposalContent.GetDescription())
	default:
		return nil
	}
}
