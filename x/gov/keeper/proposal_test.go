package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestGetSetProposal(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	app.GovKeeper.SetProposal(ctx, proposal)

	gotProposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, ProposalEqual(proposal, gotProposal))
}

func TestActivateVotingPeriod(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	app.GovKeeper.ActivateVotingPeriod(ctx, proposal)

	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := app.GovKeeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())

	proposalID := types.GetProposalIDFromBytes(activeIterator.Value())
	require.Equal(t, proposalID, proposal.ProposalID)
	activeIterator.Close()
}
