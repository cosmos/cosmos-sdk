package v043

import (
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v040"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// migrateWeightedVotes migrates the ADR-037 weighted votes.
func migrateJSONWeightedVotes(oldVotes v040gov.Votes) types.Votes {
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
func MigrateJSON(oldState *v040gov.GenesisState) *types.GenesisState {
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
