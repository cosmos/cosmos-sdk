package keeper

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	_ "cosmossdk.io/x/gov"
	v1 "cosmossdk.io/x/gov/types/v1"
	stakingtypes "cosmossdk.io/x/staking/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestTallyNoOneVotes(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.ctx

	createValidators(t, f, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

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
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	err = f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	assert.NilError(t, err)

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, _ := f.govKeeper.Tally(ctx, proposal)
	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, _ := createValidators(t, f, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

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
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, _, _ := f.govKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNoWithVeto), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	valAccAddrs, _ := createValidators(t, f, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

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
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", valAccAddrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddr1, v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, valAccAddr2, v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

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

	_, err = f.stakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)
	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[4], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes == false)
	assert.Assert(t, burnDeposits == false)
	assert.Assert(t, tallyResults.Equals(v1.EmptyTallyResult()) == false)
}

func TestTallyDelgatorInherit(t *testing.T) {
	t.Parallel()

	f := initFixture(t)

	ctx := f.ctx

	addrs, vals := createValidators(t, f, []int64{5, 6, 7})

	delTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 30)
	val3, found := f.stakingKeeper.GetValidator(ctx, vals[2])
	assert.Assert(t, found)

	_, err := f.stakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	assert.NilError(t, err)

	_, err = f.stakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)
	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, err = f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.NilError(t, err)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
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

	_, err = f.stakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)
	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

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

	_, err = f.stakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)
	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

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

	_, err = f.stakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)
	consAddr, err := val2.GetConsAddr()
	assert.NilError(t, err)
	assert.NilError(t, f.stakingKeeper.Jail(ctx, sdk.ConsAddress(consAddr)))

	tp := TestProposal
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

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
	proposal, err := f.govKeeper.SubmitProposal(ctx, tp, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	assert.NilError(t, err)
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	assert.NilError(t, f.govKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := f.govKeeper.Proposals.Get(ctx, proposalID)
	assert.Assert(t, ok)
	passes, burnDeposits, tallyResults, _ := f.govKeeper.Tally(ctx, proposal)

	assert.Assert(t, passes)
	assert.Assert(t, burnDeposits == false)

	expectedYes := f.stakingKeeper.TokensFromConsensusPower(ctx, 30)
	expectedAbstain := f.stakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedNo := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	expectedNoWithVeto := f.stakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedTallyResult := v1.NewTallyResult(expectedYes, expectedAbstain, expectedNo, expectedNoWithVeto, math.ZeroInt())

	assert.Assert(t, tallyResults.Equals(expectedTallyResult))
}
