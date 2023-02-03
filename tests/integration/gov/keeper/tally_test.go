package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestTallyNoOneVotes(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	createValidators(t, ctx, app, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	createValidators(t, ctx, app, []int64{2, 5, 0})

	addrs := simtestutil.AddTestAddrsIncremental(app.BankKeeper, app.StakingKeeper, ctx, 1, sdk.NewInt(10000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	err = app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	assert.NilError(t, err)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _ := app.GovKeeper.Tally(ctx, proposal)
	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	addrs, _ := createValidators(t, ctx, app, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidators51No(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _ := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNoWithVeto), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{5, 6, 7})
	valAccAddr1, valAccAddr2 := valAccAddrs[0], valAccAddrs[1]

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddr1, v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddr2, v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorOverride(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	addrs, valAddrs := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 30)
	val1, found := app.StakingKeeper.GetValidator(ctx, valAddrs[0])
	assert.Assert(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[4], delTokens, stakingtypes.Unbonded, val1, true)
	assert.NilError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[4], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorInherit(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	addrs, vals := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 30)
	val3, found := app.StakingKeeper.GetValidator(ctx, vals[2])
	assert.Assert(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	addrs, vals := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val1, found := app.StakingKeeper.GetValidator(ctx, vals[0])
	assert.Assert(t, found)
	val2, found := app.StakingKeeper.GetValidator(ctx, vals[1])
	assert.Assert(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val1, true)
	assert.NilError(t, err)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	createValidators(t, ctx, app, []int64{25, 6, 7})

	addrs, vals := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := app.StakingKeeper.GetValidator(ctx, vals[1])
	assert.Assert(t, found)
	val3, found := app.StakingKeeper.GetValidator(ctx, vals[2])
	assert.Assert(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyJailedValidator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	addrs, valAddrs := createValidators(t, ctx, app, []int64{25, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := app.StakingKeeper.GetValidator(ctx, valAddrs[1])
	assert.Assert(t, found)
	val3, found := app.StakingKeeper.GetValidator(ctx, valAddrs[2])
	assert.Assert(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	consAddr, err := val2.GetConsAddr()
	assert.NilError(t, err)
	app.StakingKeeper.Jail(ctx, sdk.ConsAddress(consAddr.Bytes()))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyValidatorMultipleDelegations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx := f.app, f.ctx

	addrs, valAddrs := createValidators(t, ctx, app, []int64{10, 10, 10})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := app.StakingKeeper.GetValidator(ctx, valAddrs[1])
	assert.Assert(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[0], delTokens, stakingtypes.Unbonded, val2, true)
	assert.NilError(t, err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)

	expectedYes := app.StakingKeeper.TokensFromConsensusPower(ctx, 30)
	expectedAbstain := app.StakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedNo := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	expectedNoWithVeto := app.StakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedTallyResult := v1.NewTallyResult(expectedYes, expectedAbstain, expectedNo, expectedNoWithVeto)

	assert.Assert(t, tallyResults.Equals(expectedTallyResult))
}
