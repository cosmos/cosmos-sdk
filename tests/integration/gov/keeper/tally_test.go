package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestTallyNoOneVotes(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.ctx

	createValidators(t, f, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"))
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	createValidators(t, f, []int64{2, 5, 0})

	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 1, math.NewInt(10000000))

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	err = f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	assert.NilError(t, err)

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, _, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)
	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, _ := createValidators(t, f, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidators51No(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, _, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)
	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, !passes) // only validators doesn't pass quorum
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsAbstain(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)

	passes, burnDeposits, participation, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, !passes)
	assert.Assert(t, burnDeposits == false)
	assert.Equal(t, participation.String(), math.LegacyOneDec().String()) // validators vote do not count
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{5, 6, 7})
	valAccAddr1, valAccAddr2 := valAccAddrs[0], valAccAddrs[1]

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddr1, v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddr2, v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorOverride(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, valAddrs := createValidators(t, f, []int64{5, 6, 7})

	delTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 30)
	val1, found := f.stakingKeeper.GetValidator(ctx, valAddrs[0])
	assert.Assert(t, found)

	_, err := f.stakingKeeper.Delegate(ctx, addrs[4], delTokens, stakingtypes.Unbonded, val1, true)
	assert.NilError(t, err)

	f.stakingKeeper.EndBlocker(ctx)

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[4], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorInheritDisabled(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, vals := createValidators(t, f, []int64{5, 6, 7})

	delTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 30)
	val3, found := f.stakingKeeper.GetValidator(ctx, vals[2])
	assert.Assert(t, found)

	_, err := f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)

	f.stakingKeeper.EndBlocker(ctx)

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, err = f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.NilError(t, err)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, !passes) // vote not inherited, proposal not passing
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, vals := createValidators(t, f, []int64{5, 6, 7})

	delTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val1, found := f.stakingKeeper.GetValidator(ctx, vals[0])
	assert.Assert(t, found)
	val2, found := f.stakingKeeper.GetValidator(ctx, vals[1])
	assert.Assert(t, found)

	_, err := f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val1, true)
	assert.NilError(t, err)
	_, err = f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)

	f.stakingKeeper.EndBlocker(ctx)

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	createValidators(t, f, []int64{25, 6, 7})

	addrs, vals := createValidators(t, f, []int64{5, 6, 7})

	delTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := f.stakingKeeper.GetValidator(ctx, vals[1])
	assert.Assert(t, found)
	val3, found := f.stakingKeeper.GetValidator(ctx, vals[2])
	assert.Assert(t, found)

	_, err := f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)
	_, err = f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)

	f.stakingKeeper.EndBlocker(ctx)

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyJailedValidator(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, valAddrs := createValidators(t, f, []int64{25, 6, 7})

	delTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := f.stakingKeeper.GetValidator(ctx, valAddrs[1])
	assert.Assert(t, found)
	val3, found := f.stakingKeeper.GetValidator(ctx, valAddrs[2])
	assert.Assert(t, found)

	_, err := f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)
	_, err = f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)

	f.stakingKeeper.EndBlocker(ctx)

	consAddr, err := val2.GetConsAddr()
	assert.NilError(t, err)
	f.stakingKeeper.Jail(ctx, sdk.ConsAddress(consAddr))

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyValidatorMultipleDelegations(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, valAddrs := createValidators(t, f, []int64{10, 10, 10})

	delTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := f.stakingKeeper.GetValidator(ctx, valAddrs[1])
	assert.Assert(t, found)

	_, err := f.stakingKeeper.Delegate(ctx, addrs[0], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)

	expectedYes := f.stakingKeeper.TokensFromConsensusPower(ctx, 30)
	expectedAbstain := f.stakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedNo := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	expectedTallyResult := v1.NewTallyResult(expectedYes, expectedAbstain, expectedNo)

	assert.Assert(t, tallyResults.Equals(expectedTallyResult))
}

func TestTallyGovernors(t *testing.T) {
	t.Parallel()

	f := initFixture(t)
	msgSrvr := keeper.NewMsgServerImpl(f.govKeeper)

	// Create validators
	addrs, vals := createValidators(t, f, []int64{50000, 60000, 70000})

	// Create a governor
	govTokens := math.NewInt(1000_000000)
	govAddr := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, f.ctx, 1, govTokens)[0]
	val3, found := f.stakingKeeper.GetValidator(f.ctx, vals[2])
	assert.Assert(t, found)
	_, err := f.stakingKeeper.Delegate(f.ctx, govAddr, govTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)
	f.stakingKeeper.EndBlocker(f.ctx)
	_, err = msgSrvr.CreateGovernor(f.ctx, v1.NewMsgCreateGovernor(
		govAddr, v1.GovernorDescription{},
	))
	assert.NilError(t, err)

	// Create a big delegator
	delTokens := math.NewInt(1_000_000_000000)
	delAddr := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, f.ctx, 1, delTokens)[0]
	_, err = f.stakingKeeper.Delegate(f.ctx, delAddr, delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)
	f.stakingKeeper.EndBlocker(f.ctx)
	// Delegate to governor
	msgSrvr.DelegateGovernor(f.ctx, v1.NewMsgDelegateGovernor(delAddr, types.GovernorAddress(govAddr)))

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(f.ctx, tp, "", "test", "description", addrs[0])
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	f.govKeeper.SetProposal(f.ctx, proposal)

	assert.NilError(t, f.govKeeper.AddVote(f.ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(f.ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(f.ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	// governor vote
	assert.NilError(t, f.govKeeper.AddVote(f.ctx, proposalID, govAddr, v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, err = f.govKeeper.Proposals.Get(f.ctx, proposalID)
	assert.NilError(t, err)
	passes, burnDeposits, _, tallyResults, err := f.govKeeper.Tally(f.ctx, proposal)
	assert.NilError(t, err)

	assert.Assert(t, passes)
	assert.Assert(t, !burnDeposits)
	// Expect yes count to be the sum of the governor and delegator's tokens.
	// Delegator didnt vote but has delegated its voting power to the governor.
	assert.Equal(t, tallyResults.YesCount, govTokens.Add(delTokens).String())
	assert.Equal(t, tallyResults.NoCount, "180000000000")
	assert.Equal(t, tallyResults.AbstainCount, "0")
}
