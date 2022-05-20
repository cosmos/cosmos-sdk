package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestProposalVotingPeriod_String(t *testing.T) {
	pvp := types.ProposalVotingPeriod{
		ProposalType: "cosmos.params.v1beta1.ParameterChangeProposal",
		VotingPeriod: time.Hour * 24 * 2,
	}
	expected := `proposaltype: cosmos.params.v1beta1.ParameterChangeProposal
voting_period: 48h0m0s
`
	require.Equal(t, expected, pvp.String())
}

func TestProposalVotingPeriods_Equal(t *testing.T) {
	testCases := []struct {
		name     string
		input    types.ProposalVotingPeriods
		other    types.ProposalVotingPeriods
		expectEq bool
	}{
		{
			name: "equal",
			input: []types.ProposalVotingPeriod{
				{
					ProposalType: "cosmos.params.v1beta1.ParameterChangeProposal",
					VotingPeriod: time.Hour * 24 * 2,
				},
				{
					ProposalType: "cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
					VotingPeriod: time.Hour * 24,
				},
			},
			other: []types.ProposalVotingPeriod{
				{
					ProposalType: "cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
					VotingPeriod: time.Hour * 24,
				},
				{
					ProposalType: "cosmos.params.v1beta1.ParameterChangeProposal",
					VotingPeriod: time.Hour * 24 * 2,
				},
			},
			expectEq: true,
		},
		{
			name:     "empty equal",
			input:    []types.ProposalVotingPeriod{},
			other:    []types.ProposalVotingPeriod{},
			expectEq: true,
		},
		{
			name: "not equal",
			input: []types.ProposalVotingPeriod{
				{
					ProposalType: "cosmos.params.v1beta1.ParameterChangeProposal",
					VotingPeriod: time.Hour,
				},
				{
					ProposalType: "cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
					VotingPeriod: time.Hour,
				},
			},
			other: []types.ProposalVotingPeriod{
				{
					ProposalType: "cosmos.params.v1beta1.ParameterChangeProposal",
					VotingPeriod: time.Hour * 24 * 2,
				},
				{
					ProposalType: "cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
					VotingPeriod: time.Hour * 24,
				},
			},
			expectEq: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectEq, tc.input.Equal(tc.other))
		})
	}
}
