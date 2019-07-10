package v036

import (
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
)

func MigrateGovernance(initialState v034gov.GenesisState) GenesisState {
	targetGov := GenesisState{
		StartingProposalID: initialState.StartingProposalID,
		DepositParams: DepositParams{
			MinDeposit:       initialState.DepositParams.MinDeposit,
			MaxDepositPeriod: initialState.DepositParams.MaxDepositPeriod,
		},
		TallyParams: TallyParams{
			Quorum:    initialState.TallyParams.Quorum,
			Threshold: initialState.TallyParams.Threshold,
			Veto:      initialState.TallyParams.Veto,
		},
		VotingParams: VotingParams{
			VotingPeriod: initialState.VotingParams.VotingPeriod,
		},
	}

	var deposits Deposits
	for _, p := range initialState.Deposits {
		deposits = append(deposits, Deposit{
			ProposalID: p.Deposit.ProposalID,
			Amount:     p.Deposit.Amount,
			Depositor:  p.Deposit.Depositor,
		})
	}

	targetGov.Deposits = deposits

	var votes Votes
	for _, p := range initialState.Votes {
		votes = append(votes, Vote{
			ProposalID: p.Vote.ProposalID,
			Option:     VoteOption(p.Vote.Option),
			Voter:      p.Vote.Voter,
		})
	}

	targetGov.Votes = votes

	var proposals Proposals
	for _, p := range initialState.Proposals {
		proposal := Proposal{
			Content:          migrateContent(p.ProposalContent),
			ProposalID:       p.ProposalID,
			Status:           ProposalStatus(p.Status),
			FinalTallyResult: TallyResult(p.FinalTallyResult),
			SubmitTime:       p.SubmitTime,
			DepositEndTime:   p.DepositEndTime,
			TotalDeposit:     p.TotalDeposit,
			VotingStartTime:  p.VotingStartTime,
			VotingEndTime:    p.VotingEndTime,
		}

		proposals = append(proposals, proposal)
	}

	targetGov.Proposals = proposals
	return targetGov
}

func migrateContent(proposalContent v034gov.ProposalContent) (content Content) {
	switch proposalContent.ProposalType() {
	case v034gov.ProposalTypeText:
		return NewTextProposal(proposalContent.GetTitle(), proposalContent.GetDescription())
	case v034gov.ProposalTypeSoftwareUpgrade:
		return NewSoftwareUpgradeProposal(proposalContent.GetTitle(), proposalContent.GetDescription())
	default:
		return nil
	}
}
