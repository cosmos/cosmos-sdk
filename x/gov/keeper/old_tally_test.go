package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestTallyValidatorMultipleDelegations(t *testing.T) {
	ctx, _, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{10, 10, 10})

	delTokens := sdk.TokensFromConsensusPower(10)
	val2, found := sk.GetValidator(ctx, valOpAddr2)
	require.True(t, found)

	_, err := sk.Delegate(ctx, valAccAddr1, delTokens, sdk.Unbonded, val2, true)
	require.NoError(t, err)

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionNo))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionYes))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)

	expectedYes := sdk.TokensFromConsensusPower(30)
	expectedAbstain := sdk.TokensFromConsensusPower(0)
	expectedNo := sdk.TokensFromConsensusPower(10)
	expectedNoWithVeto := sdk.TokensFromConsensusPower(0)
	expectedTallyResult := types.NewTallyResult(expectedYes, expectedAbstain, expectedNo, expectedNoWithVeto)

	require.True(t, tallyResults.Equals(expectedTallyResult))
}
