package group_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	group "github.com/cosmos/cosmos-sdk/contrib/x/group"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestGroupPolicyInfoGetDecisionPolicy_MalformedPolicy(t *testing.T) {
	// Simulate malformed policy unpacking: DecisionPolicy Any contains wrong type.
	// GetDecisionPolicy should return ErrInvalidType instead of panicking.
	coin := sdk.NewInt64Coin("stake", 1)
	wrongAny, err := codectypes.NewAnyWithValue(&coin)
	require.NoError(t, err)

	policyInfo := group.GroupPolicyInfo{
		Address:        "cosmos1abc123",
		GroupId:        1,
		Admin:          "cosmos1admin",
		Version:        1,
		DecisionPolicy: wrongAny,
	}

	_, err = policyInfo.GetDecisionPolicy()
	require.Error(t, err)
	require.True(t, sdkerrors.ErrInvalidType.Is(err), "expected ErrInvalidType, got %v", err)
}

func TestThresholdDecisionPolicyValidate(t *testing.T) {
	g := group.GroupInfo{
		TotalWeight: "10",
	}
	config := group.DefaultConfig()
	testCases := []struct {
		name   string
		policy group.ThresholdDecisionPolicy
		expErr bool
	}{
		{
			"min exec period too big",
			group.ThresholdDecisionPolicy{
				Threshold: "5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod:       time.Second,
					MinExecutionPeriod: time.Hour * 24 * 30,
				},
			},
			true,
		},
		{
			"all good",
			group.ThresholdDecisionPolicy{
				Threshold: "5",
				Windows: &group.DecisionPolicyWindows{
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
	g := group.GroupInfo{}
	config := group.DefaultConfig()
	testCases := []struct {
		name   string
		policy group.PercentageDecisionPolicy
		expErr bool
	}{
		{
			"min exec period too big",
			group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod:       time.Second,
					MinExecutionPeriod: time.Hour * 24 * 30,
				},
			},
			true,
		},
		{
			"all good",
			group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
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
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"YesCount percentage == decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"4",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"YesCount percentage < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"sum percentage (YesCount + undecided votes percentage) < decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "2",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
			false,
		},
		{
			"sum percentage = decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "2",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"4",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"sum percentage > decision policy percentage",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"empty total power",
			&group.PercentageDecisionPolicy{
				Percentage: "0.5",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"0",
			time.Second * 50,
			group.DecisionPolicyResult{
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
				Threshold: "2",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"YesCount < threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold: "2",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: false,
				Final: false,
			},
			false,
		},
		{
			"YesCount == group total weight < threshold",
			&group.ThresholdDecisionPolicy{
				Threshold: "20",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "3",
				NoCount:         "0",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: true,
				Final: true,
			},
			false,
		},
		{
			"maxYesCount < threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold: "2",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "0",
				NoCount:         "1",
				AbstainCount:    "1",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
				Allow: false,
				Final: true,
			},
			false,
		},
		{
			"maxYesCount >= threshold decision policy",
			&group.ThresholdDecisionPolicy{
				Threshold: "2",
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * 100,
				},
			},
			&group.TallyResult{
				YesCount:        "0",
				NoCount:         "1",
				AbstainCount:    "0",
				NoWithVetoCount: "0",
			},
			"3",
			time.Second * 50,
			group.DecisionPolicyResult{
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
