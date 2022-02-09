package group_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"

	"github.com/stretchr/testify/require"
)

func TestAllow(t *testing.T) {
	testCases := []struct {
		name           string
		policy         *group.PercentageDecisionPolicy
		tally          *group.Tally
		totalPower     string
		votingDuration time.Duration
		expErr         bool
	}{
		{
			"YesCount percentage >= decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.Tally{
				YesCount:     "2",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			"3",
			time.Duration(time.Second * 50),
			false,
		},
		{
			"YesCount percentage < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.Tally{
				YesCount:     "1",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			"3",
			time.Duration(time.Second * 50),
			false,
		},
		{
			"sum percentage < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.Tally{
				YesCount:     "1",
				NoCount:      "2",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			"3",
			time.Duration(time.Second * 50),
			false,
		},
		{
			"sum percentage >= decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.Tally{
				YesCount:     "0",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			"3",
			time.Duration(time.Second * 50),
			false,
		},
		{
			"decision policy timeout <= voting duration",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 10,
			},
			&group.Tally{
				YesCount:     "2",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			"3",
			time.Duration(time.Second * 50),
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyResult, err := tc.policy.Allow(*tc.tally, tc.totalPower, tc.votingDuration)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			policyPercentage, err := math.NewPositiveDecFromString(tc.policy.Percentage)
			require.NoError(t, err)
			totalPower, err := math.NewNonNegativeDecFromString(tc.totalPower)
			require.NoError(t, err)
			yesCount, err := math.NewNonNegativeDecFromString(tc.tally.YesCount)
			require.NoError(t, err)
			yesPercentage, err := yesCount.Quo(totalPower)
			require.NoError(t, err)
			totalCounts, err := tc.tally.TotalCounts()
			require.NoError(t, err)
			undecided, err := math.SubNonNegative(totalPower, totalCounts)
			require.NoError(t, err)
			sum, err := yesCount.Add(undecided)
			require.NoError(t, err)
			sumPercentage, err := sum.Quo(totalPower)
			require.NoError(t, err)

			if tc.policy.Timeout <= tc.votingDuration {
				require.Equal(t, policyResult.Allow, false)
				require.Equal(t, policyResult.Final, true)
			} else if yesPercentage.Cmp(policyPercentage) >= 0 {
				require.Equal(t, policyResult.Allow, true)
				require.Equal(t, policyResult.Final, true)
			} else if sumPercentage.Cmp(policyPercentage) < 0 {
				fmt.Println(sumPercentage.Cmp(policyPercentage), totalCounts, undecided, sum, totalPower)
				require.Equal(t, policyResult.Allow, false)
				require.Equal(t, policyResult.Final, true)
			} else {
				require.Equal(t, policyResult.Allow, false)
				require.Equal(t, policyResult.Final, false)
			}
		})
	}
}
