package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetProposal(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposalID, err := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	require.NoError(t, err)
	proposal := keeper.GetProposal(ctx, proposalID)
	info := keeper.GetProposalInfo(ctx, proposalID)
	keeper.SetProposal(ctx, proposalID, proposal)
	keeper.SetProposalInfo(ctx, info)

	gotProposal := keeper.GetProposal(ctx, proposalID)
	gotInfo := keeper.GetProposalInfo(ctx, proposalID)
	require.Equal(t, proposal, gotProposal)
	require.Equal(t, info, gotInfo)
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
	proposal6, err := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	require.NoError(t, err)
	require.Equal(t, int64(6), proposal6)
}

func TestActivateVotingPeriod(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposalID, err := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)

	info := keeper.GetProposalInfo(ctx, proposalID)

	require.NoError(t, err)
	require.True(t, info.VotingStartTime.Equal(time.Time{}))
	require.True(t, keeper.ActiveInfoQueuePeek(ctx).IsEmpty())

	keeper.activateVotingPeriod(ctx, info)

	require.True(t, info.VotingStartTime.Equal(ctx.BlockHeader().Time))
	require.Equal(t, info.ProposalID, keeper.ActiveInfoQueuePeek(ctx).ProposalID)
}

func TestDeposits(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 2)
	SortAddresses(addrs)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	proposalID, err := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	require.NoError(t, err)

	fourSteak := sdk.Coins{sdk.NewInt64Coin("steak", 4)}
	fiveSteak := sdk.Coins{sdk.NewInt64Coin("steak", 5)}

	addr0Initial := keeper.ck.GetCoins(ctx, addrs[0])
	addr1Initial := keeper.ck.GetCoins(ctx, addrs[1])

	// require.True(t, addr0Initial.IsEqual(sdk.Coins{sdk.NewInt64Coin("steak", 42)}))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin("steak", 42)}, addr0Initial)

	require.True(t, keeper.GetProposalInfo(ctx, proposalID).TotalDeposit.IsEqual(sdk.Coins{}))

	// Check no deposits at beginning
	deposit, found := keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.False(t, found)
	require.True(t, keeper.GetProposalInfo(ctx, proposalID).VotingStartTime.Equal(time.Time{}))
	require.True(t, keeper.ActiveInfoQueuePeek(ctx).IsEmpty())

	// Check first deposit
	err, votingStarted := keeper.AddDeposit(ctx, proposalID, addrs[0], fourSteak)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourSteak, deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositer)
	require.Equal(t, fourSteak, keeper.GetProposalInfo(ctx, proposalID).TotalDeposit)
	require.Equal(t, addr0Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check a second deposit from same address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveSteak)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, fourSteak.Plus(fiveSteak), deposit.Amount)
	require.Equal(t, addrs[0], deposit.Depositer)
	require.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetProposalInfo(ctx, proposalID).TotalDeposit)
	require.Equal(t, addr0Initial.Minus(fourSteak).Minus(fiveSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	// Check third deposit from a new address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, addrs[1], fourSteak)
	require.Nil(t, err)
	require.True(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1], deposit.Depositer)
	require.Equal(t, fourSteak, deposit.Amount)
	require.Equal(t, fourSteak.Plus(fiveSteak).Plus(fourSteak), keeper.GetProposalInfo(ctx, proposalID).TotalDeposit)
	require.Equal(t, addr1Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[1]))

	// Check that proposal moved to voting period
	require.True(t, keeper.GetProposalInfo(ctx, proposalID).VotingStartTime.Equal(ctx.BlockHeader().Time))
	require.NotNil(t, keeper.ActiveInfoQueuePeek(ctx))
	require.Equal(t, proposalID, keeper.ActiveInfoQueuePeek(ctx).ProposalID)

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
	depositsIterator.Close()

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

	proposalID, err := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	require.NoError(t, err)

	info := keeper.GetProposalInfo(ctx, proposalID)

	info.Status = StatusVotingPeriod
	keeper.SetProposalInfo(ctx, info)

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
	votesIterator.Close()
}

func TestInfoQueues(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	mapp.InitChainer(ctx, abci.RequestInitChain{})

	require.True(t, keeper.InactiveInfoQueuePeek(ctx).IsEmpty())
	require.True(t, keeper.ActiveInfoQueuePeek(ctx).IsEmpty())

	// create test proposals
	proposal, err := keeper.NewTextProposal(ctx, "Test", "description", ProposalTypeText)
	require.NoError(t, err)
	proposal2, err := keeper.NewTextProposal(ctx, "Test2", "description", ProposalTypeText)
	require.NoError(t, err)
	proposal3, err := keeper.NewTextProposal(ctx, "Test3", "description", ProposalTypeText)
	require.NoError(t, err)
	proposal4, err := keeper.NewTextProposal(ctx, "Test4", "description", ProposalTypeText)
	require.NoError(t, err)

	// test pushing to inactive proposal queue
	keeper.InactiveInfoQueuePush(ctx, proposal)
	keeper.InactiveInfoQueuePush(ctx, proposal2)
	keeper.InactiveInfoQueuePush(ctx, proposal3)
	keeper.InactiveInfoQueuePush(ctx, proposal4)

	// test peeking and popping from inactive proposal queue
	require.Equal(t, keeper.InactiveInfoQueuePeek(ctx).ProposalID, proposal)
	require.Equal(t, keeper.InactiveInfoQueuePop(ctx).ProposalID, proposal)
	require.Equal(t, keeper.InactiveInfoQueuePeek(ctx).ProposalID, proposal2)
	require.Equal(t, keeper.InactiveInfoQueuePop(ctx).ProposalID, proposal2)
	require.Equal(t, keeper.InactiveInfoQueuePeek(ctx).ProposalID, proposal3)
	require.Equal(t, keeper.InactiveInfoQueuePop(ctx).ProposalID, proposal3)
	require.Equal(t, keeper.InactiveInfoQueuePeek(ctx).ProposalID, proposal4)
	require.Equal(t, keeper.InactiveInfoQueuePop(ctx).ProposalID, proposal4)

	// test pushing to active proposal queue
	keeper.ActiveInfoQueuePush(ctx, proposal)
	keeper.ActiveInfoQueuePush(ctx, proposal2)
	keeper.ActiveInfoQueuePush(ctx, proposal3)
	keeper.ActiveInfoQueuePush(ctx, proposal4)

	// test peeking and popping from active proposal queue
	require.Equal(t, keeper.ActiveInfoQueuePeek(ctx).ProposalID, proposal)
	require.Equal(t, keeper.ActiveInfoQueuePop(ctx).ProposalID, proposal)
	require.Equal(t, keeper.ActiveInfoQueuePeek(ctx).ProposalID, proposal2)
	require.Equal(t, keeper.ActiveInfoQueuePop(ctx).ProposalID, proposal2)
	require.Equal(t, keeper.ActiveInfoQueuePeek(ctx).ProposalID, proposal3)
	require.Equal(t, keeper.ActiveInfoQueuePop(ctx).ProposalID, proposal3)
	require.Equal(t, keeper.ActiveInfoQueuePeek(ctx).ProposalID, proposal4)
	require.Equal(t, keeper.ActiveInfoQueuePop(ctx).ProposalID, proposal4)
}
