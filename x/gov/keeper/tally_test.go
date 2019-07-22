package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestTallyNoOneVotes(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.True(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{2, 5, 0})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := keeper.Tally(ctx, proposal)
	require.False(t, passes)
	require.True(t, burnDeposits)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionYes))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidators51No(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionNo))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionNo))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionYes))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionNoWithVeto))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))

}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionAbstain))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionNo))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionYes))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionAbstain))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionNo))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 6, 7})

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionYes))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionNo))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyDelgatorOverride(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 6, 7})

	delTokens := sdk.TokensFromConsensusPower(30)
	val1, found := sk.GetValidator(ctx, valOpAddr1)
	require.True(t, found)

	_, err := sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val1, true)
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

func TestTallyDelgatorInherit(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
	createValidators(ctx, sk, []int64{5, 6, 7})

	delTokens := sdk.TokensFromConsensusPower(30)
	val3, found := sk.GetValidator(ctx, valOpAddr3)
	require.True(t, found)

	_, err := sk.Delegate(ctx, TestAddrs[0], delTokens, sdk.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, sk)

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr1, types.OptionNo))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr2, types.OptionNo))
	require.NoError(t, keeper.AddVote(ctx, proposalID, valAccAddr3, types.OptionYes))

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
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
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
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
	ctx, _, keeper, sk, _ := createTestInput(t, false, 100)
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

	sk.Jail(ctx, sdk.ConsAddress(val2.ConsPubKey.Address()))

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
