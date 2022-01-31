package v046

import (
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

// MigrateJSON accepts exported v0.43 x/gov genesis state and migrates it to
// v0.46 x/gov genesis state. The migration includes:
//
// - Updating everything to v1beta2.
// - Migrating proposals to be Msg-based.
func MigrateJSON(oldState *v1beta1.GenesisState) (*v1beta2.GenesisState, error) {
	newProps, err := convertProposals(oldState.Proposals)
	if err != nil {
		return nil, err
	}

	depParams, votingParms, tallyParams := convertDepParams(oldState.DepositParams), convertVotingParams(oldState.VotingParams), convertTallyParams(oldState.TallyParams)

	return &v1beta2.GenesisState{
		StartingProposalId: oldState.StartingProposalId,
		Deposits:           convertDeposits(oldState.Deposits),
		Votes:              convertVotes(oldState.Votes),
		Proposals:          newProps,
		DepositParams:      &depParams,
		VotingParams:       &votingParms,
		TallyParams:        &tallyParams,
	}, nil
}
