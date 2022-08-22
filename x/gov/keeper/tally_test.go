package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestTallyNoOneVotes(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.True(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{2, 5, 0})

	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 1, sdk.NewInt(10000000))

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	err = govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.Nil(t, err)

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := govKeeper.Tally(ctx, proposal)
	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	addrs, _ := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidators51No(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	valAccAddrs, _ := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	valAccAddrs, _ := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	valAccAddrs, _ := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNoWithVeto), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	valAccAddrs, _ := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	valAccAddrs, _ := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	valAccAddrs, _ := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 6, 7})
	valAccAddr1, valAccAddr2 := valAccAddrs[0], valAccAddrs[1]

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddr1, v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, valAccAddr2, v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorOverride(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	addrs, valAddrs := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 6, 7})

	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 30)
	val1, found := stakingKeeper.GetValidator(ctx, valAddrs[0])
	require.True(t, found)

	_, err := stakingKeeper.Delegate(ctx, addrs[4], delTokens, stakingtypes.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[4], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorInherit(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	addrs, vals := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 6, 7})

	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 30)
	val3, found := stakingKeeper.GetValidator(ctx, vals[2])
	require.True(t, found)

	_, err := stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	addrs, vals := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 6, 7})

	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val1, found := stakingKeeper.GetValidator(ctx, vals[0])
	require.True(t, found)
	val2, found := stakingKeeper.GetValidator(ctx, vals[1])
	require.True(t, found)

	_, err := stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val1, true)
	require.NoError(t, err)
	_, err = stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{25, 6, 7})

	addrs, vals := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{5, 6, 7})

	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := stakingKeeper.GetValidator(ctx, vals[1])
	require.True(t, found)
	val3, found := stakingKeeper.GetValidator(ctx, vals[2])
	require.True(t, found)

	_, err := stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)
	_, err = stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyJailedValidator(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	addrs, valAddrs := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{25, 6, 7})

	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := stakingKeeper.GetValidator(ctx, valAddrs[1])
	require.True(t, found)
	val3, found := stakingKeeper.GetValidator(ctx, valAddrs[2])
	require.True(t, found)

	_, err := stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)
	_, err = stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	consAddr, err := val2.GetConsAddr()
	require.NoError(t, err)
	stakingKeeper.Jail(ctx, sdk.ConsAddress(consAddr.Bytes()))

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyValidatorMultipleDelegations(t *testing.T) {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)

	addrs, valAddrs := createValidators(t, acctKeeper, bankKeeper, stakingKeeper, ctx, []int64{10, 10, 10})

	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := stakingKeeper.GetValidator(ctx, valAddrs[1])
	require.True(t, found)

	_, err := stakingKeeper.Delegate(ctx, addrs[0], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	govKeeper.SetProposal(ctx, proposal)

	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := govKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)

	expectedYes := stakingKeeper.TokensFromConsensusPower(ctx, 30)
	expectedAbstain := stakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedNo := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	expectedNoWithVeto := stakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedTallyResult := v1.NewTallyResult(expectedYes, expectedAbstain, expectedNo, expectedNoWithVeto)

	require.True(t, tallyResults.Equals(expectedTallyResult))
}
