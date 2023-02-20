package v1_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func TestProposalStatus_Format(t *testing.T) {
	statusDepositPeriod, _ := v1.ProposalStatusFromString("PROPOSAL_STATUS_DEPOSIT_PERIOD")
	tests := []struct {
		pt                   v1.ProposalStatus
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
	testProposal := v1beta1.NewTextProposal("Proposal", "testing proposal")
	msgContent, err := v1.NewLegacyContent(testProposal, "cosmos1govacct")
	require.NoError(t, err)
	proposal, err := v1.NewProposal([]sdk.Msg{msgContent}, 1, time.Now(), time.Now(), "", "title", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), false)
	require.NoError(t, err)

	require.NotPanics(t, func() { _ = proposal.String() })
	require.NotEmpty(t, proposal.String())
}

func TestProposalSetExpedited(t *testing.T) {
	const startExpedited = false
	proposal, err := v1.NewProposal([]sdk.Msg{}, 1, time.Now(), time.Now(), "", "title", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), startExpedited)
	require.NoError(t, err)
	require.Equal(t, startExpedited, proposal.Expedited)

	proposal, err = v1.NewProposal([]sdk.Msg{}, 1, time.Now(), time.Now(), "", "title", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), !startExpedited)
	require.NoError(t, err)
	require.Equal(t, !startExpedited, proposal.Expedited)
}

func TestProposalGetMinDepositFromParams(t *testing.T) {
	testcases := []struct {
		expedited          bool
		expectedMinDeposit math.Int
	}{
		{
			expedited:          true,
			expectedMinDeposit: v1.DefaultMinExpeditedDepositTokens,
		},
		{
			expedited:          false,
			expectedMinDeposit: v1.DefaultMinDepositTokens,
		},
	}

	for _, tc := range testcases {
		proposal, err := v1.NewProposal([]sdk.Msg{}, 1, time.Now(), time.Now(), "", "title", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), tc.expedited)
		require.NoError(t, err)

		actualMinDeposit := proposal.GetMinDepositFromParams(v1.DefaultParams())

		require.Equal(t, 1, len(actualMinDeposit))
		require.Equal(t, sdk.DefaultBondDenom, actualMinDeposit[0].Denom)
		require.Equal(t, tc.expectedMinDeposit, actualMinDeposit[0].Amount)
	}
}
