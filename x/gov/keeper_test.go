package gov

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetProposal(t *testing.T) {

	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", "Text")
	proposalID := proposal.GetProposalID()
	keeper.SetProposal(ctx, proposal)

	gotProposal := keeper.GetProposal(ctx, proposalID)

	assert.Equal(t, proposal.GetProposalID(), gotProposal.GetProposalID())
	assert.Equal(t, proposal.GetTitle(), gotProposal.GetTitle())
	assert.Equal(t, proposal.GetDescription(), gotProposal.GetDescription())
	assert.Equal(t, proposal.GetProposalType(), gotProposal.GetProposalType())
	assert.Equal(t, proposal.GetSubmitBlock(), gotProposal.GetSubmitBlock())
	assert.True(t, proposal.GetTotalDeposit().IsEqual(gotProposal.GetTotalDeposit()))
	assert.Equal(t, proposal.GetSubmitBlock(), gotProposal.GetSubmitBlock())
}

func TestIncrementProposalNumber(t *testing.T) {

	ctx, _, keeper := createTestInput(t, false, 100)

	keeper.NewTextProposal(ctx, "Test", "description", "Text")
	keeper.NewTextProposal(ctx, "Test", "description", "Text")
	keeper.NewTextProposal(ctx, "Test", "description", "Text")
	keeper.NewTextProposal(ctx, "Test", "description", "Text")
	keeper.NewTextProposal(ctx, "Test", "description", "Text")
	proposal6 := keeper.NewTextProposal(ctx, "Test", "description", "Text")

	assert.Equal(t, int64(6), proposal6.GetProposalID())
}

func TestActivateVotingPeriod(t *testing.T) {

	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", "Text")

	assert.Equal(t, int64(-1), proposal.GetVotingStartBlock())
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	keeper.activateVotingPeriod(ctx, proposal)

	assert.Equal(t, proposal.GetVotingStartBlock(), ctx.BlockHeight())
	assert.Equal(t, proposal.GetProposalID(), keeper.ActiveProposalQueuePeek(ctx).GetProposalID())
}

func TestDeposits(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", "Text")
	proposalID := proposal.GetProposalID()

	fourSteak := sdk.Coins{sdk.Coin{"steak", 4}}
	fiveSteak := sdk.Coins{sdk.Coin{"steak", 5}}

	addr0Initial := keeper.ck.GetCoins(ctx, addrs[0])
	addr1Initial := keeper.ck.GetCoins(ctx, addrs[1])

	assert.True(t, proposal.GetTotalDeposit().IsEqual(sdk.Coins{}))

	deposit, found := keeper.GetDeposit(ctx, proposalID, addrs[1])
	assert.False(t, found)
	assert.Equal(t, keeper.GetProposal(ctx, proposalID).GetVotingStartBlock(), int64(-1))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	err := keeper.AddDeposit(ctx, proposalID, addrs[0], fourSteak)
	assert.Nil(t, err)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	assert.True(t, found)
	assert.Equal(t, fourSteak, deposit.Amount)
	assert.Equal(t, addrs[0], deposit.Depositer)
	assert.Equal(t, fourSteak, keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	assert.Equal(t, addr0Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	err = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveSteak)
	assert.Nil(t, err)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[0])
	assert.True(t, found)
	assert.Equal(t, fourSteak.Plus(fiveSteak), deposit.Amount)
	assert.Equal(t, addrs[0], deposit.Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	assert.Equal(t, addr0Initial.Minus(fourSteak).Minus(fiveSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	err = keeper.AddDeposit(ctx, proposalID, addrs[1], fourSteak)
	assert.Nil(t, err)
	deposit, found = keeper.GetDeposit(ctx, proposalID, addrs[1])
	assert.True(t, found)
	assert.Equal(t, addrs[1], deposit.Depositer)
	assert.Equal(t, fourSteak, deposit.Amount)
	assert.Equal(t, fourSteak.Plus(fiveSteak).Plus(fourSteak), keeper.GetProposal(ctx, proposalID).GetTotalDeposit())
	assert.Equal(t, addr1Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[1]))

	assert.Equal(t, ctx.BlockHeight(), keeper.GetProposal(ctx, proposalID).GetVotingStartBlock())
	assert.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	assert.Equal(t, proposalID, keeper.ActiveProposalQueuePeek(ctx).GetProposalID())

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
	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewTextProposal(ctx, "Test", "description", "Text")
	proposalID := proposal.GetProposalID()

	proposal.SetStatus(StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	keeper.AddVote(ctx, proposalID, addrs[0], "Abstain")
	vote, found := keeper.GetVote(ctx, proposalID, addrs[0])
	assert.True(t, found)
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "Abstain", vote.Option)

	keeper.AddVote(ctx, proposalID, addrs[0], "Yes")
	vote, found = keeper.GetVote(ctx, proposalID, addrs[0])
	assert.True(t, found)
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "Yes", vote.Option)

	keeper.AddVote(ctx, proposalID, addrs[1], "NoWithVeto")
	vote, found = keeper.GetVote(ctx, proposalID, addrs[1])
	assert.True(t, found)
	assert.Equal(t, addrs[1], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "NoWithVeto", vote.Option)

	votesIterator := keeper.GetVotes(ctx, proposalID)
	assert.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
	assert.True(t, votesIterator.Valid())
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "Yes", vote.Option)
	votesIterator.Next()
	assert.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
	assert.True(t, votesIterator.Valid())
	assert.Equal(t, addrs[1], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "NoWithVeto", vote.Option)
	votesIterator.Next()
	assert.False(t, votesIterator.Valid())
}

func TestProposalQueues(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)

	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	proposal := keeper.NewTextProposal(ctx, "Test", "description", "Text")
	proposal2 := keeper.NewTextProposal(ctx, "Test2", "description", "Text")
	proposal3 := keeper.NewTextProposal(ctx, "Test3", "description", "Text")
	proposal4 := keeper.NewTextProposal(ctx, "Test4", "description", "Text")

	keeper.InactiveProposalQueuePush(ctx, proposal)
	keeper.InactiveProposalQueuePush(ctx, proposal2)
	keeper.InactiveProposalQueuePush(ctx, proposal3)
	keeper.InactiveProposalQueuePush(ctx, proposal4)

	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).GetProposalID(), proposal4.GetProposalID())
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).GetProposalID(), proposal4.GetProposalID())

	keeper.ActiveProposalQueuePush(ctx, proposal)
	keeper.ActiveProposalQueuePush(ctx, proposal2)
	keeper.ActiveProposalQueuePush(ctx, proposal3)
	keeper.ActiveProposalQueuePush(ctx, proposal4)

	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal2.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal3.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).GetProposalID(), proposal4.GetProposalID())
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).GetProposalID(), proposal4.GetProposalID())
}
