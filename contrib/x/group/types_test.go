package group_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	group2 "github.com/cosmos/cosmos-sdk/contrib/x/group"
)

func TestThresholdDecisionPolicyValidate(t *testing.T) {
	g := group2.GroupInfo{
		TotalWeight: "10",
	}
	config := group2.DefaultConfig()
	testCases := []struct {
		name   string
		policy group2.ThresholdDecisionPolicy
		expErr bool
	}{
		{
			"min exec period too big",
			group2.ThresholdDecisionPolicy{
				Threshold: "5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod:       time.Second,
					MinExecutionPeriod: time.Hour * 24 * 30,
				},
			},
			true,
		},
		{
			"all good",
			group2.ThresholdDecisionPolicy{
				Threshold: "5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod:       time.Hour,
					MinExecutionPeriod: time.Hour * 24,
				},
			},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Validate(g, config)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPercentageDecisionPolicyValidate(t *testing.T) {
	g := group2.GroupInfo{}
	config := group2.DefaultConfig()
	testCases := []struct {
		name   string
		policy group2.PercentageDecisionPolicy
		expErr bool
	}{
		{
			"min exec period too big",
			group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod:       time.Second,
					MinExecutionPeriod: time.Hour * 24 * 30,
				},
			},
			true,
		},
		{
			"all good",
			group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod:       time.Hour,
					MinExecutionPeriod: time.Hour * 24,
				},
			},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Validate(g, config)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPercentageDecisionPolicyAllow(t *testing.T) {
	testCases := []struct {
		name           string
		policy         *group2.PercentageDecisionPolicy
		tally          *group2.TallyResult
		totalPower     string
		votingDuration time.Duration
		result         group2.DecisionPolicyResult
		expErr         bool
	}{
		{
			"YesCount percentage > decision policy percentage",
			&group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"YesCount percentage == decision policy percentage",
			&group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"4",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"YesCount percentage < decision policy percentage",
			&group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"sum percentage (YesCount + undecided votes percentage) < decision policy percentage",
			&group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "1",
				NoCount:         "2",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
			false,
		},
		{
			"sum percentage = decision policy percentage",
			&group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "1",
				NoCount:         "2",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"4",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"sum percentage > decision policy percentage",
			&group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"empty total power",
			&group2.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"0",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyResult, err := tc.policy.Allow(*tc.tally, tc.totalPower)
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
		policy         *group2.ThresholdDecisionPolicy
		tally          *group2.TallyResult
		totalPower     string
		votingDuration time.Duration
		result         group2.DecisionPolicyResult
		expErr         bool
	}{
		{
			"YesCount >= threshold decision policy",
			&group2.ThresholdDecisionPolicy{
				Threshold: "2",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"YesCount < threshold decision policy",
			&group2.ThresholdDecisionPolicy{
				Threshold: "2",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"YesCount == group total weight < threshold",
			&group2.ThresholdDecisionPolicy{
				Threshold: "20",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "3",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"maxYesCount < threshold decision policy",
			&group2.ThresholdDecisionPolicy{
				Threshold: "2",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "0",
				NoCount:         "1",
				AbstainCount:    "1",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
			false,
		},
		{
			"maxYesCount >= threshold decision policy",
			&group2.ThresholdDecisionPolicy{
				Threshold: "2",
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group2.TallyResult{
				YesCount:        "0",
				NoCount:         "1",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group2.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policyResult, err := tc.policy.Allow(*tc.tally, tc.totalPower)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.result, policyResult)
			}
		})
	}
}
