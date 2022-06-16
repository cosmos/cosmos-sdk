package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/stretchr/testify/require"
)

func TestTallyParamsGetThreshold(t *testing.T) {
	testcases := []struct {
		name          string
		tallyParams   types.TallyParams
		expectedValue sdk.Dec
		isExpedited   bool
	}{
		{
			name:          "default expedited",
			tallyParams:   types.DefaultTallyParams(),
			expectedValue: sdk.NewDecWithPrec(667, 3),
			isExpedited:   true,
		},
		{
			name:          "default not expedited",
			tallyParams:   types.DefaultTallyParams(),
			expectedValue: sdk.NewDecWithPrec(5, 1),
			isExpedited:   false,
		},
		{
			name:          "custom expedited",
			tallyParams:   types.NewTallyParams(types.DefaultQuorum, types.DefaultThreshold, sdk.NewDecWithPrec(777, 3), types.DefaultVetoThreshold),
			expectedValue: sdk.NewDecWithPrec(777, 3),
			isExpedited:   true,
		},
		{
			name:          "default not expedited",
			tallyParams:   types.NewTallyParams(types.DefaultQuorum, sdk.NewDecWithPrec(6, 1), types.DefaultExpeditedThreshold, types.DefaultVetoThreshold),
			expectedValue: sdk.NewDecWithPrec(6, 1),
			isExpedited:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedValue, tc.tallyParams.GetThreshold(tc.isExpedited))
		})
	}
}

func TestVotingParamsGetVotingTime(t *testing.T) {
	testcases := []struct {
		name          string
		votingParams  types.VotingParams
		expectedValue time.Duration
		isExpedited   bool
	}{
		{
			name:          "default expedited",
			votingParams:  types.DefaultVotingParams(),
			expectedValue: types.DefaultExpeditedPeriod,
			isExpedited:   true,
		},
		{
			name:          "default not expedited",
			votingParams:  types.DefaultVotingParams(),
			expectedValue: types.DefaultPeriod,
			isExpedited:   false,
		},
		{
			name:          "custom expedited",
			votingParams:  types.NewVotingParams(types.DefaultPeriod, time.Hour, types.DefaultProposalVotingPeriods),
			expectedValue: time.Hour,
			isExpedited:   true,
		},
		{
			name:          "default not expedited",
			votingParams:  types.NewVotingParams(time.Hour, types.DefaultExpeditedPeriod, types.DefaultProposalVotingPeriods),
			expectedValue: time.Hour,
			isExpedited:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedValue, tc.votingParams.GetVotingPeriod(tc.isExpedited), tc.name)
		})
	}
}

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
