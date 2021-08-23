package v043

import (
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// migrateWeightedVotes migrates the ADR-037 weighted votes.
func migrateJSONWeightedVotes(oldVotes types.Votes) types.Votes {
	newVotes := make(types.Votes, len(oldVotes))
	for i, oldVote := range oldVotes {
		newVotes[i] = migrateVote(oldVote)
	}

	return newVotes
}

// MigrateJSON accepts exported v0.40 x/gov genesis state and migrates it to
// v0.43 x/gov genesis state. The migration includes:
//
// - Gov weighted votes.
func MigrateJSON(oldState *types.GenesisState) *types.GenesisState {
	return &types.GenesisState{
		StartingProposalId: oldState.StartingProposalId,
		Deposits:           oldState.Deposits,
		Votes:              migrateJSONWeightedVotes(oldState.Votes),
		Proposals:          oldState.Proposals,
		DepositParams:      oldState.DepositParams,
		VotingParams:       oldState.VotingParams,
		TallyParams:        oldState.TallyParams,
	}
}
