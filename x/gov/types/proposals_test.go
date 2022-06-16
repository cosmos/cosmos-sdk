package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

func TestProposalStatus_Format(t *testing.T) {
	statusDepositPeriod, _ := types.ProposalStatusFromString("PROPOSAL_STATUS_DEPOSIT_PERIOD")
	tests := []struct {
		pt                   types.ProposalStatus
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

func TestProposalSetIsExpedited(t *testing.T) {
	const startIsExpedited = false
	testProposal := types.NewTextProposal("test", "description")

	proposal, err := types.NewProposal(testProposal, 1, time.Now(), time.Now(), startIsExpedited)
	require.NoError(t, err)
	require.Equal(t, startIsExpedited, proposal.IsExpedited)

	proposal, err = types.NewProposal(testProposal, 1, time.Now(), time.Now(), !startIsExpedited)
	require.Equal(t, !startIsExpedited, proposal.IsExpedited)
}
