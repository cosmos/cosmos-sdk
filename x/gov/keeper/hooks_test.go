package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	minDeposit := app.GovKeeper.GetDepositParams(ctx).MinDeposit
	addrs := simapp.AddTestAddrs(app, ctx, 1, minDeposit[0].Amount)

	govHooksReceiver := MockGovHooksReceiver{}

	app.GovKeeper = *keeper.UpdateHooks(&app.GovKeeper,
		types.NewMultiGovHooks(
			&govHooksReceiver,
		),
	)

	require.False(t, govHooksReceiver.AfterProposalSubmissionValid)
	require.False(t, govHooksReceiver.AfterProposalDepositValid)
	require.False(t, govHooksReceiver.AfterProposalVoteValid)
	require.False(t, govHooksReceiver.AfterProposalInactiveValid)
	require.False(t, govHooksReceiver.AfterProposalActiveValid)

	tp := TestProposal
	_, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalSubmissionValid)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	gov.EndBlocker(ctx, app.GovKeeper)

	require.True(t, govHooksReceiver.AfterProposalInactiveValid)

	p2, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	activated, err := app.GovKeeper.AddDeposit(ctx, p2.ProposalId, addrs[0], minDeposit)
	require.True(t, activated)
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalDepositValid)

	err = app.GovKeeper.AddVote(ctx, p2.ProposalId, addrs[0], types.NewNonSplitVoteOption(types.OptionYes))
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalVoteValid)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(app.GovKeeper.GetVotingParams(ctx).VotingPeriod).Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	gov.EndBlocker(ctx, app.GovKeeper)
	require.True(t, govHooksReceiver.AfterProposalActiveValid)
}
