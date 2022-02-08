package v1beta2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
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

// TestNestedAnys tests that we can call .String() on a struct with nested Anys.
// Here, we're creating a proposal which has a Msg (1st any) with a legacy
// content (2nd any).
func TestNestedAnys(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/10965
	t.Skip()
	testProposal := v1beta1.NewTextProposal("Proposal", "testing proposal")
	msgContent, err := v1beta2.NewLegacyContent(testProposal, "cosmos1govacct")
	require.NoError(t, err)
	proposal, err := v1beta2.NewProposal([]sdk.Msg{msgContent}, 1, nil, time.Now(), time.Now())
	require.NoError(t, err)

	require.Equal(t, "TODO Fix panic here", proposal.String())
}
