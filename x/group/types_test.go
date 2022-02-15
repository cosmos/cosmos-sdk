package group_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/group"

	"github.com/stretchr/testify/require"
)

func TestPercentageDecisionPolicyAllow(t *testing.T) {
	testCases := []struct {
		name           string
		policy         *group.PercentageDecisionPolicy
		tally          *group.TallyResult
		totalPower     string
		votingDuration time.Duration
		result         group.DecisionPolicyResult
	}{
		{
			"YesCount percentage > decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
		},
		{
			"YesCount percentage == decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"4",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
		},
		{
			"YesCount percentage < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
		},
		{
			"sum percentage (YesCount + undecided votes percentage) < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "2",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
		},
		{
			"sum percentage = decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "2",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"4",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
		},
		{
			"sum percentage > decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
		},
		{
			"decision policy timeout <= voting duration",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Timeout:    time.Second * 10,
			},
			&group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyResult, err := tc.policy.Allow(*tc.tally, tc.totalPower, tc.votingDuration)
			require.NoError(t, err)
			require.Equal(t, tc.result, policyResult)
		})
	}
}

func TestThresholdDecisionPolicyAllow(t *testing.T) {
	testCases := []struct {
		name           string
		policy         *group.ThresholdDecisionPolicy
		tally          *group.TallyResult
		totalPower     string
		votingDuration time.Duration
		result         group.DecisionPolicyResult
	}{
		{
			"YesCount >= threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold: "3",
				Timeout:   time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "3",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
		},
		{
			"YesCount < threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold: "3",
				Timeout:   time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
		},
		{
			"sum votes < threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold: "3",
				Timeout:   time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"2",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
		},
		{
			"sum votes >= threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold: "3",
				Timeout:   time.Second * 100,
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
		},
		{
			"decision policy timeout <= voting duration",
			&group.ThresholdDecisionPolicy{
				Threshold: "3",
				Timeout:   time.Second * 10,
			},
			&group.TallyResult{
				YesCount:        "3",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyResult, err := tc.policy.Allow(*tc.tally, tc.totalPower, tc.votingDuration)
			require.NoError(t, err)
			require.Equal(t, tc.result, policyResult)
		})
	}
}
