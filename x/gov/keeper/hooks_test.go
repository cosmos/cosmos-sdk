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
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

var _ types.GovHooks = &MockGovHooksReceiver{}

// GovHooks event hooks for governance proposal object (noalias)
type MockGovHooksReceiver struct {
	AfterProposalSubmissionValid        bool
	AfterProposalDepositValid           bool
	AfterProposalVoteValid              bool
	AfterProposalFailedMinDepositValid  bool
	AfterProposalVotingPeriodEndedValid bool
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
func (h *MockGovHooksReceiver) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
	h.AfterProposalFailedMinDepositValid = true
}
func (h *MockGovHooksReceiver) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	h.AfterProposalVotingPeriodEndedValid = true
}

func TestHooks(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	minDeposit := app.GovKeeper.GetDepositParams(ctx).MinDeposit
	addrs := simapp.AddTestAddrs(app, ctx, 1, minDeposit[0].Amount)

	govHooksReceiver := MockGovHooksReceiver{}

	keeper.UnsafeSetHooks(
		&app.GovKeeper, types.NewMultiGovHooks(&govHooksReceiver),
	)

	require.False(t, govHooksReceiver.AfterProposalSubmissionValid)
	require.False(t, govHooksReceiver.AfterProposalDepositValid)
	require.False(t, govHooksReceiver.AfterProposalVoteValid)
	require.False(t, govHooksReceiver.AfterProposalFailedMinDepositValid)
	require.False(t, govHooksReceiver.AfterProposalVotingPeriodEndedValid)

	tp := TestProposal
	_, err := app.GovKeeper.SubmitProposal(ctx, tp, nil)
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalSubmissionValid)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	gov.EndBlocker(ctx, app.GovKeeper)

	require.True(t, govHooksReceiver.AfterProposalFailedMinDepositValid)

	p2, err := app.GovKeeper.SubmitProposal(ctx, tp, nil)
	require.NoError(t, err)

	activated, err := app.GovKeeper.AddDeposit(ctx, p2.Id, addrs[0], minDeposit)
	require.True(t, activated)
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalDepositValid)

	err = app.GovKeeper.AddVote(ctx, p2.Id, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes))
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalVoteValid)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*app.GovKeeper.GetVotingParams(ctx).VotingPeriod).Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	gov.EndBlocker(ctx, app.GovKeeper)
	require.True(t, govHooksReceiver.AfterProposalVotingPeriodEndedValid)
}
