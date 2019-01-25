package gov

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	pubkeys = []crypto.PubKey{ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey()}

	testDescription   = staking.NewDescription("T", "E", "S", "T")
	testCommissionMsg = staking.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func createValidators(t *testing.T, stakingHandler sdk.Handler, ctx sdk.Context, addrs []sdk.ValAddress, coinAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {
		valCreateMsg := staking.NewMsgCreateValidator(
			addrs[i], pubkeys[i], sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, coinAmt[i]), testDescription, testCommissionMsg,
		)

		res := stakingHandler(ctx, valCreateMsg)
		require.True(t, res.IsOK())
	}
}

func TestTallyNoOneVotes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 5})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.True(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{2, 5})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))
	require.False(t, passes)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 5})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidators51No(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)

	passes, _ := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNoWithVeto)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{6, 6, 7})
	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorOverride(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6, 7})
	staking.EndBlocker(ctx, sk)

	delegator1Msg := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[2]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 30))
	stakingHandler(ctx, delegator1Msg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[3], OptionNo)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorInherit(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6, 7})
	staking.EndBlocker(ctx, sk)

	delegator1Msg := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[2]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 30))
	stakingHandler(ctx, delegator1Msg)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorMultipleOverride(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 6, 7})
	staking.EndBlocker(ctx, sk)

	delegator1Msg := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[2]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 10))
	stakingHandler(ctx, delegator1Msg)
	delegator1Msg2 := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[1]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 10))
	stakingHandler(ctx, delegator1Msg2)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[3], OptionNo)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyDelgatorMultipleInherit(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	val1CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(addrs[0]), ed25519.GenPrivKey().PubKey(), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 25), testDescription, testCommissionMsg,
	)
	stakingHandler(ctx, val1CreateMsg)

	val2CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(addrs[1]), ed25519.GenPrivKey().PubKey(), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 6), testDescription, testCommissionMsg,
	)
	stakingHandler(ctx, val2CreateMsg)

	val3CreateMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(addrs[2]), ed25519.GenPrivKey().PubKey(), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 7), testDescription, testCommissionMsg,
	)
	stakingHandler(ctx, val3CreateMsg)

	delegator1Msg := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[2]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 10))
	stakingHandler(ctx, delegator1Msg)

	delegator1Msg2 := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[1]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 10))
	stakingHandler(ctx, delegator1Msg2)

	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.False(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyJailedValidator(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{25, 6, 7})
	staking.EndBlocker(ctx, sk)

	delegator1Msg := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[2]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 10))
	stakingHandler(ctx, delegator1Msg)

	delegator1Msg2 := staking.NewMsgDelegate(addrs[3], sdk.ValAddress(addrs[1]), sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 10))
	stakingHandler(ctx, delegator1Msg2)

	val2, found := sk.GetValidator(ctx, sdk.ValAddress(addrs[1]))
	require.True(t, found)
	sk.Jail(ctx, sdk.ConsAddress(val2.ConsPubKey.Address()))

	staking.EndBlocker(ctx, sk)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	err := keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	passes, tallyResults := tally(ctx, keeper, keeper.GetProposal(ctx, proposalID))

	require.True(t, passes)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}
