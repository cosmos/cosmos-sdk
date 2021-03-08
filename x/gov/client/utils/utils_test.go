package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/gov/client/utils"
)

func TestNormalizeProposalStatus(t *testing.T) {
	type args struct {
		status string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"invalid", args{"unknown"}, "unknown"},
		{"deposit_period", args{"deposit_period"}, "PROPOSAL_STATUS_DEPOSIT_PERIOD"},
		{"DepositPeriod", args{"DepositPeriod"}, "PROPOSAL_STATUS_DEPOSIT_PERIOD"},
		{"voting_period", args{"deposit_period"}, "PROPOSAL_STATUS_DEPOSIT_PERIOD"},
		{"VotingPeriod", args{"DepositPeriod"}, "PROPOSAL_STATUS_DEPOSIT_PERIOD"},
		{"passed", args{"passed"}, "PROPOSAL_STATUS_PASSED"},
		{"Passed", args{"Passed"}, "PROPOSAL_STATUS_PASSED"},
		{"Rejected", args{"Rejected"}, "PROPOSAL_STATUS_REJECTED"},
		{"rejected", args{"rejected"}, "PROPOSAL_STATUS_REJECTED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, utils.NormalizeProposalStatus(tt.args.status))
		})
	}
}
