package v4

import (
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func MigrateJSON(oldState *v1.GenesisState) (*v1.GenesisState, error) {
	params := v1.NewParams(
		oldState.DepositParams.MinDeposit,
		*oldState.DepositParams.MaxDepositPeriod,
		*oldState.VotingParams.VotingPeriod,
		oldState.TallyParams.Quorum,
		oldState.TallyParams.Threshold,
		oldState.TallyParams.VetoThreshold,
		v1.DefaultParams().MinInitialDepositRatio,
	)

	return &v1.GenesisState{
		StartingProposalId: oldState.StartingProposalId,
		Deposits:           oldState.Deposits,
		Votes:              oldState.Votes,
		Proposals:          oldState.Proposals,
		Params:             &params,
	}, nil
}
