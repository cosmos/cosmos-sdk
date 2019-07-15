package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestTallyNoOneVotes(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
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
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
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
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionYes)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidators51No(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionNoWithVeto)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))

}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionYes)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyDelgatorOverride(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	delTokens := sdk.TokensFromConsensusPower(30)
	delegator1Msg := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[3], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyDelgatorInherit(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	delTokens := sdk.TokensFromConsensusPower(30)
	delegator1Msg := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionYes)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	delTokens := sdk.TokensFromConsensusPower(10)
	delegator1Msg := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)
	delegator1Msg2 := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[1]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg2)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[3], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)
	stakingHandler := staking.NewHandler(input.sk)

	valTokens1 := sdk.TokensFromConsensusPower(25)
	val1CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(TestAddrs[0]), ed25519.GenPrivKey().PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens1), testDescription, testCommissionRates, sdk.OneInt(),
	)
	stakingHandler(ctx, val1CreateMsg)

	valTokens2 := sdk.TokensFromConsensusPower(6)
	val2CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(TestAddrs[1]), ed25519.GenPrivKey().PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens2), testDescription, testCommissionRates, sdk.OneInt(),
	)
	stakingHandler(ctx, val2CreateMsg)

	valTokens3 := sdk.TokensFromConsensusPower(7)
	val3CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(TestAddrs[2]), ed25519.GenPrivKey().PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens3), testDescription, testCommissionRates, sdk.OneInt(),
	)
	stakingHandler(ctx, val3CreateMsg)

	delTokens := sdk.TokensFromConsensusPower(10)
	delegator1Msg := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	delegator1Msg2 := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[1]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg2)

	staking.EndBlocker(ctx, input.sk)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyJailedValidator(t *testing.T) {
	ctx, _, keeper, _ := createTestInput(t, false, 100)

	delTokens := sdk.TokensFromConsensusPower(10)
	delegator1Msg := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	delegator1Msg2 := staking.NewMsgDelegate(TestAddrs[3], sdk.ValAddress(TestAddrs[1]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg2)

	val2, found := input.sk.GetValidator(ctx, sdk.ValAddress(TestAddrs[1]))
	require.True(t, found)
	input.sk.Jail(ctx, sdk.ConsAddress(val2.ConsPubKey.Address()))

	staking.EndBlocker(ctx, input.sk)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = types.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	err = keeper.AddVote(ctx, proposalID, TestAddrs[0], types.OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[1], types.OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, TestAddrs[2], types.OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}
