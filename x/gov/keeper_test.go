package gov

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetProposal(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	proposalID := proposal.GetProposalID()
	keeper.SetProposal(ctx, proposal)

	gotProposal := keeper.GetProposal(ctx, proposalID)
	assert.True(t, ProposalEqual(proposal, gotProposal))
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

	assert.Equal(t, int64(6), proposal6.GetProposalID())
}

func TestActivateVotingPeriod(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposal := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	assert.Equal(t, int64(-1), proposal.GetVotingStartBlock())
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	keeper.activateVotingPeriod(ctx, proposal)

	assert.Equal(t, proposal.GetVotingStartBlock(), ctx.BlockHeight())
	assert.Equal(t, proposal.GetProposalID(), keeper.ActiveProposalQueuePeek(ctx).GetProposalID())
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

	// assert.True(t, addr0Initial.IsEqual(sdk.Coins{sdk.NewCoin("steak", 42)}))
	assert.Equal(t, sdk.Coins{sdk.NewCoin("steak", 42)}, addr0Initial)

	assert.True(t, proposal.GetTotalDeposit().IsEqual(sdk.Coins{}))

	// Check no deposits at beginning
	deposit, found := keeper.GetDeposit(ctx, proposalID, addrs[1])
	assert.False(t, found)
	assert.Equal(t, keeper.GetProposal(ctx, proposalID).GetVotingStartBlock(), int64(-1))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	// Check first deposit
	err, votingStarted := keeper.AddDeposit(ctx, proposalID, addrs[0], fourSteak)
	assert.Nil(t, err)
	assert.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	assert.True(t, found)
	assert.Equal(t, fourSteak, deposit.Amount)
	assert.Equal(t, addrs[0], deposit.Depositer)
	assert.Equal(t, fourSteak, keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	assert.Equal(t, addr0Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check a second deposit from same address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveSteak)
	assert.Nil(t, err)
	assert.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	assert.True(t, found)
	assert.Equal(t, fourSteak.Plus(fiveSteak), deposit.Amount)
	assert.Equal(t, addrs[0], deposit.Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	assert.Equal(t, addr0Initial.Minus(fourSteak).Minus(fiveSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check third deposit from a new address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[1], fourSteak)
	assert.Nil(t, err)
	assert.True(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	assert.True(t, found)
	assert.Equal(t, addrs[1], deposit.Depositer)
	assert.Equal(t, fourSteak, deposit.Amount)
	assert.Equal(t, fourSteak.Plus(fiveSteak).Plus(fourSteak), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	assert.Equal(t, addr1Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[1]))

	// Check that proposal moved to voting period
	assert.Equal(t, ctx.BlockHeight(), keeper.GetProposal(ctx, proposalID).GetVotingStartBlock())
	assert.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	assert.Equal(t, proposalID, keeper.ActiveProposalQueuePeek(ctx).GetProposalID())

	// Test deposit iterator
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	assert.True(t, depositsIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &deposit)
	assert.Equal(t, addrs[0], deposit.Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak), deposit.Amount)
	depositsIterator.Next()
	keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &deposit)
	assert.Equal(t, addrs[1], deposit.Depositer)
	assert.Equal(t, fourSteak, deposit.Amount)
	depositsIterator.Next()
	assert.False(t, depositsIterator.Valid())

	// Test Refund Deposits
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	assert.True(t, found)
	assert.Equal(t, fourSteak, deposit.Amount)
	keeper.RefundDeposits(ctx, proposalID)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	assert.False(t, found)
	assert.Equal(t, addr0Initial, keeper.ck.GetCoins(ctx, addrs[0]))
	assert.Equal(t, addr1Initial, keeper.ck.GetCoins(ctx, addrs[1]))

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
	assert.True(t, found)
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, OptionAbstain, vote.Option)

	// Test change of vote
	keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	vote, found = keeper.GetVote(ctx, proposalID, addrs[0])
	assert.True(t, found)
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, OptionYes, vote.Option)

	// Test second vote
	keeper.AddVote(ctx, proposalID, addrs[1], OptionNoWithVeto)
	vote, found = keeper.GetVote(ctx, proposalID, addrs[1])
	assert.True(t, found)
	assert.Equal(t, addrs[1], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, OptionNoWithVeto, vote.Option)

	// Test vote iterator
	votesIterator := keeper.GetVotes(ctx, proposalID)
	assert.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
	assert.True(t, votesIterator.Valid())
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, OptionYes, vote.Option)
	votesIterator.Next()
	assert.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
	assert.True(t, votesIterator.Valid())
	assert.Equal(t, addrs[1], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, OptionNoWithVeto, vote.Option)
	votesIterator.Next()
	assert.False(t, votesIterator.Valid())
}

func TestProposalQueues(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	mapp.InitChainer(ctx, abci.RequestInitChain{})

	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

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
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal4.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal4.GetProposalID())

	// test pushing to active proposal queue
	keeper.ActiveProposalQueuePush(ctx, proposal)
	keeper.ActiveProposalQueuePush(ctx, proposal2)
	keeper.ActiveProposalQueuePush(ctx, proposal3)
	keeper.ActiveProposalQueuePush(ctx, proposal4)

	// test peeking and popping from active proposal queue
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal4.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal4.GetProposalID())
}
