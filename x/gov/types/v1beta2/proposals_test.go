package v1beta2_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
	"github.com/stretchr/testify/require"
)

func TestProposalStatus_Format(t *testing.T) {
	statusDepositPeriod, _ := v1beta2.ProposalStatusFromString("PROPOSAL_STATUS_DEPOSIT_PERIOD")
	tests := []struct {
		pt                   v1beta2.ProposalStatus
		sprintFArgs          string
		expectedStringOutput string
	}{
		{statusDepositPeriod, "%s", "PROPOSAL_STATUS_DEPOSIT_PERIOD"},
		{statusDepositPeriod, "%v", "1"},
	}
	for _, tt := range tests {
		got := fmt.Sprintf(tt.sprintFArgs, tt.pt)
		require.Equal(t, tt.expectedStringOutput, got)
	}
}
