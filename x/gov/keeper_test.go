package gov

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// overwrite defaults for testing
func init() {
	defaultMinDeposit = 10
	defaultMaxDepositPeriod = 200
	defaultVotingPeriod = 200
}

func TestGetSetProposal(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	keeper.SetProposal(ctx, proposal)

	gotProposal := keeper.GetProposal(ctx, proposalID)
	require.True(t, ProposalEqual(proposal, gotProposal))
}

func TestIncrementProposalNumber(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposal6 := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	require.Equal(t, int64(6), proposal6.GetProposalID())
}

func TestActivateVotingPeriod(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	require.Equal(t, int64(-1), proposal.GetVotingStartBlock())
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	keeper.activateVotingPeriod(ctx, proposal)

	require.Equal(t, proposal.GetVotingStartBlock(), ctx.BlockHeight())
	require.Equal(t, proposal.GetProposalID(), keeper.ActiveProposalQueuePeek(ctx).GetProposalID())
}

func TestDeposits(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2)
	SortAddresses(addrs)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()

	fourSteak := sdk.Coins{sdk.NewCoin("steak", 4)}
	fiveSteak := sdk.Coins{sdk.NewCoin("steak", 5)}

	addr0Initial := keeper.ck.GetCoins(ctx, addrs[0])
	addr1Initial := keeper.ck.GetCoins(ctx, addrs[1])

	// require.True(t, addr0Initial.IsEqual(sdk.Coins{sdk.NewCoin("steak", 42)}))
	require.Equal(t, sdk.Coins{sdk.NewCoin("steak", 42)}, addr0Initial)

	require.True(t, proposal.GetTotalDeposit().IsEqual(sdk.Coins{}))

	// Check no deposits at beginning
	deposit, found := keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.False(t, found)
	require.Equal(t, keeper.GetProposal(ctx, proposalID).GetVotingStartBlock(), int64(-1))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	// Check first deposit
	err, votingStarted := keeper.AddDeposit(ctx, proposalID, addrs[0], fourSteak)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourSteak, deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositer)
	require.Equal(t, fourSteak, keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	require.Equal(t, addr0Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check a second deposit from same address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveSteak)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourSteak.Plus(fiveSteak), deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositer)
	require.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	require.Equal(t, addr0Initial.Minus(fourSteak).Minus(fiveSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check third deposit from a new address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[1], fourSteak)
	require.Nil(t, err)
	require.True(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1], deposit.Depositer)
	require.Equal(t, fourSteak, deposit.Amount)
	require.Equal(t, fourSteak.Plus(fiveSteak).Plus(fourSteak), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	require.Equal(t, addr1Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[1]))

	// Check that proposal moved to voting period
	require.Equal(t, ctx.BlockHeight(), keeper.GetProposal(ctx, proposalID).GetVotingStartBlock())
	require.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.Equal(t, proposalID, keeper.ActiveProposalQueuePeek(ctx).GetProposalID())

	// Test deposit iterator
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &deposit)
	require.Equal(t, addrs[0], deposit.Depositer)
	require.Equal(t, fourSteak.Plus(fiveSteak), deposit.Amount)
	depositsIterator.Next()
	keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &deposit)
	require.Equal(t, addrs[1], deposit.Depositer)
	require.Equal(t, fourSteak, deposit.Amount)
	depositsIterator.Next()
	require.False(t, depositsIterator.Valid())

	// Test Refund Deposits
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, fourSteak, deposit.Amount)
	keeper.RefundDeposits(ctx, proposalID)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.False(t, found)
	require.Equal(t, addr0Initial, keeper.ck.GetCoins(ctx, addrs[0]))
	require.Equal(t, addr1Initial, keeper.ck.GetCoins(ctx, addrs[1]))

}

func TestVotes(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2)
	SortAddresses(addrs)
	mapp.BeginBlock(abci.RequestBeginBlock{})
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
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionYes, vote.Option)
	votesIterator.Next()
	require.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
	require.True(t, votesIterator.Valid())
	require.Equal(t, addrs[1], vote.Voter)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, OptionNoWithVeto, vote.Option)
	votesIterator.Next()
	require.False(t, votesIterator.Valid())
}

func TestProposalQueues(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	mapp.InitChainer(ctx, abci.RequestInitChain{})

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	// create test proposals
	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposal2 := keeper.NewTextProposal(ctx, "Test2", "description", ProposalTypeText)
	proposal3 := keeper.NewTextProposal(ctx, "Test3", "description", ProposalTypeText)
	proposal4 := keeper.NewTextProposal(ctx, "Test4", "description", ProposalTypeText)

	// test pushing to inactive proposal queue
	keeper.InactiveProposalQueuePush(ctx, proposal)
	keeper.InactiveProposalQueuePush(ctx, proposal2)
	keeper.InactiveProposalQueuePush(ctx, proposal3)
	keeper.InactiveProposalQueuePush(ctx, proposal4)

	// test peeking and popping from inactive proposal queue
	require.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal.GetProposalID())
	require.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal.GetProposalID())
	require.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal2.GetProposalID())
	require.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal2.GetProposalID())
	require.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal3.GetProposalID())
	require.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal3.GetProposalID())
	require.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal4.GetProposalID())
	require.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal4.GetProposalID())

	// test pushing to active proposal queue
	keeper.ActiveProposalQueuePush(ctx, proposal)
	keeper.ActiveProposalQueuePush(ctx, proposal2)
	keeper.ActiveProposalQueuePush(ctx, proposal3)
	keeper.ActiveProposalQueuePush(ctx, proposal4)

	// test peeking and popping from active proposal queue
	require.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal.GetProposalID())
	require.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal.GetProposalID())
	require.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal2.GetProposalID())
	require.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal2.GetProposalID())
	require.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal3.GetProposalID())
	require.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal3.GetProposalID())
	require.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal4.GetProposalID())
	require.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal4.GetProposalID())
}
