package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	ctx, _, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 6, 7})

	delTokens := sdk.TokensFromConsensusPower(10)
	val1, found := sk.GetValidator(ctx, valOpAddr1)
	require.True(t, found)
	val2, found := sk.GetValidator(ctx, valOpAddr2)
	require.True(t, found)

	_, err := sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)
	_, err = sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val2, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, sk)

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionNo))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	ctx, _, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{25, 6, 7})

	delTokens := sdk.TokensFromConsensusPower(10)
	val2, found := sk.GetValidator(ctx, valOpAddr2)
	require.True(t, found)
	val3, found := sk.GetValidator(ctx, valOpAddr3)
	require.True(t, found)

	_, err := sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val2, true)
	require.NoError(t, err)
	_, err = sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, sk)

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionNo))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionNo))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyJailedValidator(t *testing.T) {
	ctx, _, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{25, 6, 7})

	delTokens := sdk.TokensFromConsensusPower(10)
	val2, found := sk.GetValidator(ctx, valOpAddr2)
	require.True(t, found)
	val3, found := sk.GetValidator(ctx, valOpAddr3)
	require.True(t, found)

	_, err := sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val2, true)
	require.NoError(t, err)
	_, err = sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, sk)

	sk.Jail(ctx, sdk.ConsAddress(val2.GetConsPubKey().Address()))

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionNo))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionNo))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

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
