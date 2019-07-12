package gov

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestTallyNoOneVotes(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:2]))
	for i, addr := range input.addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 5})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.True(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:2]))
	for i, addr := range input.addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{2, 5})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := tally(ctx, input.keeper, proposal)
	require.False(t, passes)
	require.True(t, burnDeposits)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:2]))
	for i, addr := range input.addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 5})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionYes)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidators51No(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:2]))
	for i, addr := range input.addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionNoWithVeto)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))

}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionNo)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionYes)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorOverride(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	delTokens := sdk.TokensFromConsensusPower(30)
	delegator1Msg := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[3], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorInherit(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	delTokens := sdk.TokensFromConsensusPower(30)
	delegator1Msg := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionNo)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionNo)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionYes)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	delTokens := sdk.TokensFromConsensusPower(10)
	delegator1Msg := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)
	delegator1Msg2 := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[1]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg2)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[3], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valTokens1 := sdk.TokensFromConsensusPower(25)
	val1CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(input.addrs[0]), ed25519.GenPrivKey().PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens1), testDescription, testCommissionRates, sdk.OneInt(),
	)
	stakingHandler(ctx, val1CreateMsg)

	valTokens2 := sdk.TokensFromConsensusPower(6)
	val2CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(input.addrs[1]), ed25519.GenPrivKey().PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens2), testDescription, testCommissionRates, sdk.OneInt(),
	)
	stakingHandler(ctx, val2CreateMsg)

	valTokens3 := sdk.TokensFromConsensusPower(7)
	val3CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(input.addrs[2]), ed25519.GenPrivKey().PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens3), testDescription, testCommissionRates, sdk.OneInt(),
	)
	stakingHandler(ctx, val3CreateMsg)

	delTokens := sdk.TokensFromConsensusPower(10)
	delegator1Msg := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	delegator1Msg2 := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[1]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg2)

	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionNo)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyJailedValidator(t *testing.T) {
	input := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(input.sk)

	valAddrs := make([]sdk.ValAddress, len(input.addrs[:3]))
	for i, addr := range input.addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{25, 6, 7})
	staking.EndBlocker(ctx, input.sk)

	delTokens := sdk.TokensFromConsensusPower(10)
	delegator1Msg := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[2]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg)

	delegator1Msg2 := staking.NewMsgDelegate(input.addrs[3], sdk.ValAddress(input.addrs[1]), sdk.NewCoin(sdk.DefaultBondDenom, delTokens))
	stakingHandler(ctx, delegator1Msg2)

	val2, found := input.sk.GetValidator(ctx, sdk.ValAddress(input.addrs[1]))
	require.True(t, found)
	input.sk.Jail(ctx, sdk.ConsAddress(val2.ConsPubKey.Address()))

	staking.EndBlocker(ctx, input.sk)

	tp := testProposal()
	proposal, err := input.keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	input.keeper.SetProposal(ctx, proposal)

	err = input.keeper.AddVote(ctx, proposalID, input.addrs[0], OptionYes)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[1], OptionNo)
	require.Nil(t, err)
	err = input.keeper.AddVote(ctx, proposalID, input.addrs[2], OptionNo)
	require.Nil(t, err)

	proposal, ok := input.keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := tally(ctx, input.keeper, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}
