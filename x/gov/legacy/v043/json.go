package v043

import (
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v040"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func migrateJSONDeposits(oldDeposits v040gov.Deposits) types.Deposits {
	newDeps := make(types.Deposits, len(oldDeposits))
	for i, oldDep := range oldDeposits {
		newDeps[i] = types.Deposit(oldDep)
	}

	return newDeps
}

// migrateWeightedVotes migrates the ADR-037 weighted votes.
func migrateJSONWeightedVotes(oldVotes v040gov.Votes) types.Votes {
	newVotes := make(types.Votes, len(oldVotes))
	for i, oldVote := range oldVotes {
		newVotes[i] = migrateVote(oldVote)
	}

	return newVotes
}

func migrateJSONProposals(oldProposals v040gov.Proposals) types.Proposals {
	newProps := make(types.Proposals, len(oldProposals))
	for i, oldProp := range oldProposals {
		newProps[i] = types.Proposal{
			ProposalId:       oldProp.ProposalId,
			Content:          oldProp.Content,
			Status:           types.ProposalStatus(oldProp.Status),
			FinalTallyResult: types.TallyResult(oldProp.FinalTallyResult),
			SubmitTime:       oldProp.SubmitTime,
			DepositEndTime:   oldProp.DepositEndTime,
			TotalDeposit:     oldProp.TotalDeposit,
			VotingStartTime:  oldProp.VotingStartTime,
			VotingEndTime:    oldProp.VotingEndTime,
		}
	}

	return newProps
}

// MigrateJSON accepts exported v0.40 x/gov genesis state and migrates it to
// v0.43 x/gov genesis state. The migration includes:
//
// - Gov weighted votes.
func MigrateJSON(oldState *v040gov.GenesisState) *types.GenesisState {
	return &types.GenesisState{
		StartingProposalId: oldState.StartingProposalId,
		Deposits:           migrateJSONDeposits(oldState.Deposits),
		Votes:              migrateJSONWeightedVotes(oldState.Votes),
		Proposals:          migrateJSONProposals(oldState.Proposals),
		DepositParams:      types.DepositParams(oldState.DepositParams),
		VotingParams:       types.VotingParams(oldState.VotingParams),
		TallyParams:        types.TallyParams(oldState.TallyParams),
	}
}
