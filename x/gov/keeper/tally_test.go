package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestTallyNoOneVotes(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	createValidators(t, ctx, app, []int64{5, 5, 5})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.True(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	createValidators(t, ctx, app, []int64{2, 5, 0})

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(10000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	err = app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.Nil(t, err)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := app.GovKeeper.Tally(ctx, proposal)
	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs, _ := createValidators(t, ctx, app, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidators51No(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{5, 6, 0})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNoWithVeto), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{6, 6, 7})

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[0], v1.NewNonSplitVoteOption(v1.OptionAbstain), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valAccAddrs, _ := createValidators(t, ctx, app, []int64{5, 6, 7})
	valAccAddr1, valAccAddr2 := valAccAddrs[0], valAccAddrs[1]

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddr1, v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, valAccAddr2, v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorOverride(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs, valAddrs := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 30)
	val1, found := app.StakingKeeper.GetValidator(ctx, valAddrs[0])
	require.True(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[4], delTokens, stakingtypes.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[4], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorInherit(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs, vals := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 30)
	val3, found := app.StakingKeeper.GetValidator(ctx, vals[2])
	require.True(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs, vals := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val1, found := app.StakingKeeper.GetValidator(ctx, vals[0])
	require.True(t, found)
	val2, found := app.StakingKeeper.GetValidator(ctx, vals[1])
	require.True(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val1, true)
	require.NoError(t, err)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[3], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	createValidators(t, ctx, app, []int64{25, 6, 7})

	addrs, vals := createValidators(t, ctx, app, []int64{5, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := app.StakingKeeper.GetValidator(ctx, vals[1])
	require.True(t, found)
	val3, found := app.StakingKeeper.GetValidator(ctx, vals[2])
	require.True(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyJailedValidator(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs, valAddrs := createValidators(t, ctx, app, []int64{25, 6, 7})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := app.StakingKeeper.GetValidator(ctx, valAddrs[1])
	require.True(t, found)
	val3, found := app.StakingKeeper.GetValidator(ctx, valAddrs[2])
	require.True(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)
	_, err = app.StakingKeeper.Delegate(ctx, addrs[3], delTokens, stakingtypes.Unbonded, val3, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	consAddr, err := val2.GetConsAddr()
	require.NoError(t, err)
	app.StakingKeeper.Jail(ctx, sdk.ConsAddress(consAddr.Bytes()))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionNo), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(v1.EmptyTallyResult()))
}

func TestTallyValidatorMultipleDelegations(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs, valAddrs := createValidators(t, ctx, app, []int64{10, 10, 10})

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	val2, found := app.StakingKeeper.GetValidator(ctx, valAddrs[1])
	require.True(t, found)

	_, err := app.StakingKeeper.Delegate(ctx, addrs[0], delTokens, stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposalID := proposal.Id
	proposal.Status = v1.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], v1.NewNonSplitVoteOption(v1.OptionNo), ""))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)

	expectedYes := app.StakingKeeper.TokensFromConsensusPower(ctx, 30)
	expectedAbstain := app.StakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedNo := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	expectedNoWithVeto := app.StakingKeeper.TokensFromConsensusPower(ctx, 0)
	expectedTallyResult := v1.NewTallyResult(expectedYes, expectedAbstain, expectedNo, expectedNoWithVeto)

	require.True(t, tallyResults.Equals(expectedTallyResult))
}
