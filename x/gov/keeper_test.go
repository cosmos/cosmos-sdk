package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetProposal(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	keeper.SetProposal(ctx, proposal)

	gotProposal := keeper.GetProposal(ctx, proposalID)
	require.True(t, ProposalEqual(proposal, gotProposal))
}

func TestIncrementProposalNumber(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposal6 := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	require.Equal(t, uint64(6), proposal6.GetProposalID())
}

func TestActivateVotingPeriod(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	require.True(t, proposal.GetVotingStartTime().Equal(time.Time{}))

	keeper.activateVotingPeriod(ctx, proposal)

	require.True(t, proposal.GetVotingStartTime().Equal(ctx.BlockHeader().Time))

	activeIterator := keeper.ActiveProposalQueueIterator(ctx, proposal.GetVotingEndTime())
	require.True(t, activeIterator.Valid())
	var proposalID uint64
	keeper.cdc.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.GetProposalID())
	activeIterator.Close()
}

func TestDeposits(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(addrs)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()

	fourStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(4)))
	fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(5)))

	addr0Initial := keeper.ck.GetCoins(ctx, addrs[0])
	addr1Initial := keeper.ck.GetCoins(ctx, addrs[1])

	expTokens := sdk.TokensFromTendermintPower(42)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, expTokens)), addr0Initial)
	require.True(t, proposal.GetTotalDeposit().IsEqual(sdk.NewCoins()))

	// Check no deposits at beginning
	deposit, found := keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.False(t, found)
	require.True(t, keeper.GetProposal(ctx, proposalID).GetVotingStartTime().Equal(time.Time{}))

	// Check first deposit
	err, votingStarted := keeper.AddDeposit(ctx, proposalID, addrs[0], fourStake)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositor)
	require.Equal(t, fourStake, keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	require.Equal(t, addr0Initial.Sub(fourStake), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check a second deposit from same address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveStake)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourStake.Add(fiveStake), deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositor)
	require.Equal(t, fourStake.Add(fiveStake), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	require.Equal(t, addr0Initial.Sub(fourStake).Sub(fiveStake), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check third deposit from a new address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[1], fourStake)
	require.Nil(t, err)
	require.True(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1], deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	require.Equal(t, fourStake.Add(fiveStake).Add(fourStake), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	require.Equal(t, addr1Initial.Sub(fourStake), keeper.ck.GetCoins(ctx, addrs[1]))

	// Check that proposal moved to voting period
	require.True(t, keeper.GetProposal(ctx, proposalID).GetVotingStartTime().Equal(ctx.BlockHeader().Time))

	// Test deposit iterator
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, addrs[0], deposit.Depositor)
	require.Equal(t, fourStake.Add(fiveStake), deposit.Amount)
	depositsIterator.Next()
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, addrs[1], deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	depositsIterator.Next()
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()

	// Test Refund Deposits
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	keeper.RefundDeposits(ctx, proposalID)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.False(t, found)
	require.Equal(t, addr0Initial, keeper.ck.GetCoins(ctx, addrs[0]))
	require.Equal(t, addr1Initial, keeper.ck.GetCoins(ctx, addrs[1]))

}

func TestVotes(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(addrs)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()

	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	// Test first vote
	keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	vote, found := keeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionAbstain, vote.Option)

	// Test change of vote
	keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	vote, found = keeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionYes, vote.Option)

	// Test second vote
	keeper.AddVote(ctx, proposalID, addrs[1], OptionNoWithVeto)
	vote, found = keeper.GetVote(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionNoWithVeto, vote.Option)

	// Test vote iterator
	votesIterator := keeper.GetVotes(ctx, proposalID)
	require.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionYes, vote.Option)
	votesIterator.Next()
	require.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, addrs[1], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionNoWithVeto, vote.Option)
	votesIterator.Next()
	require.False(t, votesIterator.Valid())
	votesIterator.Close()
}

func TestProposalQueues(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	mapp.InitChainer(ctx, abci.RequestInitChain{})

	// create test proposals
	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	inactiveIterator := keeper.InactiveProposalQueueIterator(ctx, proposal.GetDepositEndTime())
	require.True(t, inactiveIterator.Valid())
	var proposalID uint64
	keeper.cdc.UnmarshalBinaryLengthPrefixed(inactiveIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.GetProposalID())
	inactiveIterator.Close()

	keeper.activateVotingPeriod(ctx, proposal)

	activeIterator := keeper.ActiveProposalQueueIterator(ctx, proposal.GetVotingEndTime())
	require.True(t, activeIterator.Valid())
	keeper.cdc.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.GetProposalID())
	activeIterator.Close()
}
