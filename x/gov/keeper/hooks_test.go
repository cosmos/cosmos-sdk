package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GovHooks event hooks for governance proposal object (noalias)
type MockGovHooksReceiver struct {
	AfterProposalSubmissionValid bool
	AfterProposalDepositValid    bool
	AfterProposalVoteValid       bool
	AfterProposalInactiveValid   bool
	AfterProposalActiveValid     bool
}

func (h *MockGovHooksReceiver) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	h.AfterProposalSubmissionValid = true
}

func (h *MockGovHooksReceiver) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	h.AfterProposalDepositValid = true
}

func (h *MockGovHooksReceiver) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	h.AfterProposalVoteValid = true
}
func (h *MockGovHooksReceiver) AfterProposalInactive(ctx sdk.Context, proposalID uint64) {
	h.AfterProposalInactiveValid = true
}
func (h *MockGovHooksReceiver) AfterProposalActive(ctx sdk.Context, proposalID uint64) {
	h.AfterProposalActiveValid = true
}

func TestHooks(t *testing.T) {
	app := simapp.Setup(false)

	govHooksReceiver := MockGovHooksReceiver{}

	app.GovKeeper = *keeper.UpdateHooks(&app.GovKeeper,
		types.NewMultiGovHooks(
			&govHooksReceiver,
		),
	)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tp := TestProposal
	_, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.True(t, govHooksReceiver.AfterProposalSubmissionValid)
}
