package v1_test

import (
	"testing"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"
)

func TestEmptyGenesis(t *testing.T) {
	state1 := v1.GenesisState{}
	require.True(t, state1.Empty())

	state2 := v1.DefaultGenesisState()
	require.False(t, state2.Empty())
}

func TestValidateGenesis(t *testing.T) {
	depositParams := v1.DefaultDepositParams()
	votingParams := v1.DefaultVotingParams()
	tallyParams := v1.DefaultTallyParams()

	testCases := []struct {
		name         string
		genesisState *v1.GenesisState
		expErr       bool
	}{
		{
			name:         "valid",
			genesisState: v1.DefaultGenesisState(),
		},
		{
			name: "invalid StartingProposalId",
			genesisState: &v1.GenesisState{
				StartingProposalId: 0,
				DepositParams:      &depositParams,
				VotingParams:       &votingParams,
				TallyParams:        &tallyParams,
			},
			expErr: true,
		},
		{
			name: "invalid TallyParams",
			genesisState: &v1.GenesisState{
				StartingProposalId: v1.DefaultStartingProposalID,
				DepositParams:      &depositParams,
				VotingParams:       &votingParams,
				TallyParams:        &v1.TallyParams{},
			},
			expErr: true,
		},
		{
			name: "invalid VotingParams",
			genesisState: &v1.GenesisState{
				StartingProposalId: v1.DefaultStartingProposalID,
				DepositParams:      &depositParams,
				VotingParams:       &v1.VotingParams{},
				TallyParams:        &tallyParams,
			},
			expErr: true,
		},
		{
			name: "invalid DepositParams",
			genesisState: &v1.GenesisState{
				StartingProposalId: v1.DefaultStartingProposalID,
				DepositParams:      &v1.DepositParams{},
				VotingParams:       &votingParams,
				TallyParams:        &tallyParams,
			},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := v1.ValidateGenesis(tc.genesisState)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
