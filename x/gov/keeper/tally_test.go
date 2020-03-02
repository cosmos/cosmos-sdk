package keeper_test

import (
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestTallyNoOneVotes(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	createValidators(ctx, app, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.True(t, tallyResults.Equals(types.EmptyTallyResult()))
}
