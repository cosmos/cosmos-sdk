package gov

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetProposal(t *testing.T) {

	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewProposal(ctx, "Test", "description", "Text")
	proposalID := proposal.ProposalID
	keeper.SetProposal(ctx, proposal)

	gotProposal := keeper.GetProposal(ctx, proposalID)

	assert.Equal(t, proposal.ProposalID, gotProposal.ProposalID)
	assert.Equal(t, proposal.Title, gotProposal.Title)
	assert.Equal(t, proposal.Description, gotProposal.Description)
	assert.Equal(t, proposal.ProposalType, gotProposal.ProposalType)
	assert.Equal(t, proposal.SubmitBlock, gotProposal.SubmitBlock)
	assert.True(t, proposal.TotalDeposit.IsEqual(gotProposal.TotalDeposit))
	assert.Equal(t, proposal.SubmitBlock, gotProposal.SubmitBlock)
}

func TestIncrementProposalNumber(t *testing.T) {

	ctx, _, keeper := createTestInput(t, false, 100)

	keeper.NewProposal(ctx, "Test", "description", "Text")
	keeper.NewProposal(ctx, "Test", "description", "Text")
	keeper.NewProposal(ctx, "Test", "description", "Text")
	keeper.NewProposal(ctx, "Test", "description", "Text")
	keeper.NewProposal(ctx, "Test", "description", "Text")
	proposal6 := keeper.NewProposal(ctx, "Test", "description", "Text")

	assert.Equal(t, proposal6.ProposalID, int64(6))
}

func TestActivateVotingPeriod(t *testing.T) {

	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewProposal(ctx, "Test", "description", "Text")

	assert.Equal(t, int64(-1), proposal.VotingStartBlock)
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	keeper.activateVotingPeriod(ctx, proposal)

	assert.Equal(t, proposal.VotingStartBlock, ctx.BlockHeight())
	assert.Equal(t, proposal.ProposalID, keeper.ActiveProposalQueuePeek(ctx).ProposalID)
}

func TestDeposits(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewProposal(ctx, "Test", "description", "Text")
	proposalID := proposal.ProposalID

	fourSteak := sdk.Coins{sdk.Coin{"steak", 4}}
	fiveSteak := sdk.Coins{sdk.Coin{"steak", 5}}

	addr0Initial := keeper.ck.GetCoins(ctx, addrs[0])
	addr1Initial := keeper.ck.GetCoins(ctx, addrs[1])

	assert.True(t, proposal.TotalDeposit.IsEqual(sdk.Coins{}))
	assert.Nil(t, keeper.GetDeposit(ctx, proposal.ProposalID, addrs[0]))
	assert.Equal(t, keeper.GetProposal(ctx, proposalID).VotingStartBlock, int64(-1))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	err := keeper.AddDeposit(ctx, proposalID, addrs[0], fourSteak)
	assert.Nil(t, err)
	assert.Equal(t, fourSteak, keeper.GetDeposit(ctx, proposalID, addrs[0]).Amount)
	assert.Equal(t, addrs[0], keeper.GetDeposit(ctx, proposalID, addrs[0]).Depositer)
	assert.Equal(t, fourSteak, keeper.GetProposal(ctx, proposalID).TotalDeposit)
	assert.Equal(t, addr0Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	err = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveSteak)
	assert.Nil(t, err)
	assert.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetDeposit(ctx, proposalID, addrs[0]).Amount)
	assert.Equal(t, addrs[0], keeper.GetDeposit(ctx, proposalID, addrs[0]).Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetProposal(ctx, proposalID).TotalDeposit)
	assert.Equal(t, addr0Initial.Minus(fourSteak).Minus(fiveSteak), keeper.ck.GetCoins(ctx, addrs[0]))

	err = keeper.AddDeposit(ctx, proposalID, addrs[1], fourSteak)
	assert.Nil(t, err)
	assert.Equal(t, fourSteak, keeper.GetDeposit(ctx, proposalID, addrs[1]).Amount)
	assert.Equal(t, addrs[1], keeper.GetDeposit(ctx, proposalID, addrs[1]).Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak).Plus(fourSteak), keeper.GetProposal(ctx, proposalID).TotalDeposit)
	assert.Equal(t, addr1Initial.Minus(fourSteak), keeper.ck.GetCoins(ctx, addrs[1]))

	assert.Equal(t, ctx.BlockHeight(), keeper.GetProposal(ctx, proposalID).VotingStartBlock)
	assert.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	assert.Equal(t, proposalID, keeper.ActiveProposalQueuePeek(ctx).ProposalID)

	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	assert.True(t, depositsIterator.Valid())
	nextDeposit := Deposit{}
	keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &nextDeposit)
	assert.Equal(t, addrs[0], nextDeposit.Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak), nextDeposit.Amount)
	depositsIterator.Next()
	keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &nextDeposit)
	assert.Equal(t, addrs[1], nextDeposit.Depositer)
	assert.Equal(t, fourSteak, nextDeposit.Amount)
	depositsIterator.Next()
	assert.False(t, depositsIterator.Valid())

	assert.Equal(t, fourSteak, keeper.GetDeposit(ctx, proposalID, addrs[1]).Amount)
	keeper.RefundDeposits(ctx, proposalID)
	assert.Nil(t, keeper.GetDeposit(ctx, proposalID, addrs[1]))
	assert.Equal(t, addr0Initial, keeper.ck.GetCoins(ctx, addrs[0]))
	assert.Equal(t, addr1Initial, keeper.ck.GetCoins(ctx, addrs[1]))

}

func TestVotes(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)

	proposal := keeper.NewProposal(ctx, "Test", "description", "Text")
	proposalID := proposal.ProposalID

	proposal.Status = "VotingPeriod"
	keeper.SetProposal(ctx, proposal)

	keeper.AddVote(ctx, proposalID, addrs[0], "Abstain")
	vote := keeper.GetVote(ctx, proposalID, addrs[0])
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "Abstain", vote.Option)

	keeper.AddVote(ctx, proposalID, addrs[0], "Yes")
	vote = keeper.GetVote(ctx, proposalID, addrs[0])
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "Yes", vote.Option)

	keeper.AddVote(ctx, proposalID, addrs[1], "NoWithVeto")
	vote = keeper.GetVote(ctx, proposalID, addrs[1])
	assert.Equal(t, addrs[1], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "NoWithVeto", vote.Option)

	votesIterator := keeper.GetVotes(ctx, proposalID)
	assert.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), vote)
	assert.True(t, votesIterator.Valid())
	assert.Equal(t, addrs[0], vote.Voter)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, "Yes", vote.Option)
	votesIterator.Next()
	assert.True(t, votesIterator.Valid())
	keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), vote)
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

	proposal := keeper.NewProposal(ctx, "Test", "description", "Text")
	proposal2 := keeper.NewProposal(ctx, "Test2", "description", "Text")
	proposal3 := keeper.NewProposal(ctx, "Test3", "description", "Text")
	proposal4 := keeper.NewProposal(ctx, "Test4", "description", "Text")

	keeper.InactiveProposalQueuePush(ctx, proposal)
	keeper.InactiveProposalQueuePush(ctx, proposal2)
	keeper.InactiveProposalQueuePush(ctx, proposal3)
	keeper.InactiveProposalQueuePush(ctx, proposal4)

	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).ProposalID, proposal.ProposalID)
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).ProposalID, proposal.ProposalID)
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).ProposalID, proposal2.ProposalID)
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).ProposalID, proposal2.ProposalID)
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).ProposalID, proposal3.ProposalID)
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).ProposalID, proposal3.ProposalID)
	assert.Equal(t, keeper.InactiveProposalQueuePeek(ctx).ProposalID, proposal4.ProposalID)
	assert.Equal(t, keeper.InactiveProposalQueuePop(ctx).ProposalID, proposal4.ProposalID)

	keeper.ActiveProposalQueuePush(ctx, proposal)
	keeper.ActiveProposalQueuePush(ctx, proposal2)
	keeper.ActiveProposalQueuePush(ctx, proposal3)
	keeper.ActiveProposalQueuePush(ctx, proposal4)

	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).ProposalID, proposal.ProposalID)
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).ProposalID, proposal.ProposalID)
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).ProposalID, proposal2.ProposalID)
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).ProposalID, proposal2.ProposalID)
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).ProposalID, proposal3.ProposalID)
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).ProposalID, proposal3.ProposalID)
	assert.Equal(t, keeper.ActiveProposalQueuePeek(ctx).ProposalID, proposal4.ProposalID)
	assert.Equal(t, keeper.ActiveProposalQueuePop(ctx).ProposalID, proposal4.ProposalID)
}
