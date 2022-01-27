package v046

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

func convertDeposits(oldDeps v1beta1.Deposits) v1beta2.Deposits {
	newDeps := make([]*v1beta2.Deposit, len(oldDeps))
	for i, oldDep := range oldDeps {
		newDeps[i] = &v1beta2.Deposit{
			ProposalId: oldDep.ProposalId,
			Depositor:  oldDep.Depositor,
			Amount:     oldDep.Amount,
		}
	}

	return newDeps
}

func convertVotes(oldVotes v1beta1.Votes) v1beta2.Votes {
	newVotes := make([]*v1beta2.Vote, len(oldVotes))
	for i, oldVote := range oldVotes {
		// All oldVotes don't have the Option field anymore, as they have been
		// migrated in the v043 package.
		newWVOs := make([]*v1beta2.WeightedVoteOption, len(oldVote.Options))
		for j, oldWVO := range oldVote.Options {
			newWVOs[j] = v1beta2.NewWeightedVoteOption(v1beta2.VoteOption(oldVote.Option), oldWVO.Weight)
		}

		newVotes[i] = &v1beta2.Vote{
			ProposalId: oldVote.ProposalId,
			Voter:      oldVote.Voter,
			Options:    newWVOs,
		}
	}

	return newVotes
}

func convertDepParams(oldDepParams v1beta1.DepositParams) v1beta2.DepositParams {
	return v1beta2.DepositParams{
		MinDeposit:       oldDepParams.MinDeposit,
		MaxDepositPeriod: &oldDepParams.MaxDepositPeriod,
	}
}

func convertVotingParams(oldVoteParams v1beta1.VotingParams) v1beta2.VotingParams {
	return v1beta2.VotingParams{
		VotingPeriod: &oldVoteParams.VotingPeriod,
	}
}

func convertTallyParams(oldTallyParams v1beta1.TallyParams) v1beta2.TallyParams {
	return v1beta2.TallyParams{
		Quorum:        oldTallyParams.Quorum.String(),
		Threshold:     oldTallyParams.Threshold.String(),
		VetoThreshold: oldTallyParams.VetoThreshold.String(),
	}
}

func convertProposal(oldProp v1beta1.Proposal) (v1beta2.Proposal, error) {
	msg, err := v1beta2.NewLegacyContent(oldProp.GetContent(), authtypes.NewModuleAddress(ModuleName).String())
	if err != nil {
		return v1beta2.Proposal{}, err
	}
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return v1beta2.Proposal{}, err
	}

	return v1beta2.Proposal{
		ProposalId: oldProp.ProposalId,
		Messages:   []*codectypes.Any{msgAny},
		Status:     v1beta2.ProposalStatus(oldProp.Status),
		FinalTallyResult: &v1beta2.TallyResult{
			Yes:        oldProp.FinalTallyResult.Yes.String(),
			No:         oldProp.FinalTallyResult.No.String(),
			Abstain:    oldProp.FinalTallyResult.Abstain.String(),
			NoWithVeto: oldProp.FinalTallyResult.NoWithVeto.String(),
		},
		SubmitTime:      &oldProp.SubmitTime,
		DepositEndTime:  &oldProp.DepositEndTime,
		TotalDeposit:    oldProp.TotalDeposit,
		VotingStartTime: &oldProp.VotingStartTime,
		VotingEndTime:   &oldProp.VotingEndTime,
	}, nil
}

func convertProposals(oldProps v1beta1.Proposals) (v1beta2.Proposals, error) {
	newProps := make([]*v1beta2.Proposal, len(oldProps))
	for i, oldProp := range oldProps {
		p, err := convertProposal(oldProp)
		if err != nil {
			return nil, err
		}

		newProps[i] = &p
	}

	return newProps, nil
}
