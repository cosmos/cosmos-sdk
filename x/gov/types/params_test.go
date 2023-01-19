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
			votingParams:  types.NewVotingParams(types.DefaultPeriod, time.Hour),
			expectedValue: time.Hour,
			isExpedited:   true,
		},
		{
			name:          "default not expedited",
			votingParams:  types.NewVotingParams(time.Hour, types.DefaultExpeditedPeriod),
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
