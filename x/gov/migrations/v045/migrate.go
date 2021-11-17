package v045

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func Migrate(oldGovState v1beta1.GenesisState, govAcc sdk.AccAddress) *types.GenesisState {
	return &types.GenesisState{
		Deposits:      migrateDeposits(oldGovState.Deposits),
		Votes:         migrateVotes(oldGovState.Votes),
		Proposals:     migrateProposals(oldGovState.Proposals, govAcc),
		DepositParams: types.DepositParams(oldGovState.DepositParams),
		VotingParams:  types.VotingParams(oldGovState.VotingParams),
		TallyParams:   types.TallyParams(oldGovState.TallyParams),
	}
}

func migrateProposals(oldProposals v1beta1.Proposals, govAcc sdk.AccAddress) types.Proposals {
	newProposals := make(types.Proposals, len(oldProposals))
	for idx, proposal := range oldProposals {
		content := proposal.GetContent()
		if content == nil {
			continue
		}

		msgs, err := sdktx.SetMsgs(migrateContent(content, govAcc))
		if err != nil {
			panic(fmt.Sprintf("failed to marshal proposal msgs: %v", err))
		}

		newProposals[idx] = types.Proposal{
			ProposalId:       proposal.ProposalId,
			Messages:         msgs,
			Status:           types.ProposalStatus(proposal.Status),
			FinalTallyResult: types.TallyResult(proposal.FinalTallyResult),
			SubmitTime:       proposal.SubmitTime,
			DepositEndTime:   proposal.DepositEndTime,
			VotingStartTime:  proposal.VotingStartTime,
			VotingEndTime:    proposal.VotingEndTime,
		}
	}
	return newProposals
}

func migrateContent(content v1beta1.Content, govAcc sdk.AccAddress) []sdk.Msg {
	switch content.ProposalType() {
	case v1beta1.ProposalTypeText:
		return []sdk.Msg{}
	// TODO: enter the other proposal content types
	default:
		// NOTE: If a network is using a unique content type that isn't recognisable then it will not be possible to migrate it to the new proposal type. The best thing to do in this situation, rather than silently ignore it, is to convert it to a signal proposal
		return []sdk.Msg{}
	}
}

func migrateVotes(oldVotes v1beta1.Votes) types.Votes {
	votes := make(types.Votes, len(oldVotes))
	for idx, oldVote := range oldVotes {
		votes[idx] = migrateVote(oldVote)
	}
	return votes
}

func migrateDeposits(oldDeposits v1beta1.Deposits) types.Deposits {
	deposits := make(types.Deposits, len(oldDeposits))
	for idx, oldDeposit := range oldDeposits {
		deposits[idx] = types.Deposit(oldDeposit)
	}
	return deposits
}

func migrateVote(oldVote v1beta1.Vote) types.Vote {
	return types.Vote{
		ProposalId: oldVote.ProposalId,
		Voter:      oldVote.Voter,
		Options:    migrateWeightedVoteOptions(oldVote.Options),
	}
}

func migrateWeightedVoteOptions(oldOptions v1beta1.WeightedVoteOptions) types.WeightedVoteOptions {
	options := make(types.WeightedVoteOptions, len(oldOptions))
	for idx, oldOption := range oldOptions {
		options[idx] = types.WeightedVoteOption{
			Option: types.VoteOption(oldOption.Option),
			Weight: oldOption.Weight,
		}
	}
	return options
}
