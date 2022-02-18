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
		expErr         bool
	}{
		{
			"YesCount percentage > decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage:   "0.5",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"YesCount percentage == decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage:   "0.5",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"YesCount percentage < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage:   "0.5",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"sum percentage (YesCount + undecided votes percentage) < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage:   "0.5",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"sum percentage = decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage:   "0.5",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"sum percentage > decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage:   "0.5",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"time since submission < min execution period",
			&group.PercentageDecisionPolicy{
				Percentage:         "0.5",
				VotingPeriod:       time.Second * 10,
				MinExecutionPeriod: time.Minute,
			},
			&group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Duration(time.Second * 50),
			group.DecisionPolicyResult{},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyResult, err := tc.policy.Allow(*tc.tally, tc.totalPower, tc.votingDuration)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.result, policyResult)
			}
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
		expErr         bool
	}{
		{
			"YesCount >= threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold:    "3",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"YesCount < threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold:    "3",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"sum votes < threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold:    "3",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"sum votes >= threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold:    "3",
				VotingPeriod: time.Second * 100,
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
			false,
		},
		{
			"time since submission < min execution period",
			&group.ThresholdDecisionPolicy{
				Threshold:          "3",
				VotingPeriod:       time.Second * 10,
				MinExecutionPeriod: time.Minute,
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
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyResult, err := tc.policy.Allow(*tc.tally, tc.totalPower, tc.votingDuration)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.result, policyResult)
			}
		})
	}
}
