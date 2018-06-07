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
	proposal5 := keeper.NewProposal(ctx, "Test", "description", "Text")

	assert.Equal(t, proposal5.ProposalID, int64(5))
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

	assert.True(t, proposal.TotalDeposit.IsEqual(sdk.Coins{}))
	assert.Nil(t, keeper.GetDeposit(ctx, proposal.ProposalID, addrs[0]))
	assert.Equal(t, keeper.GetProposal(ctx, proposalID).VotingStartBlock, int64(-1))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))

	err := keeper.AddDeposit(ctx, proposalID, addrs[0], fourSteak)
	assert.Nil(t, err)
	assert.Equal(t, fourSteak, keeper.GetDeposit(ctx, proposalID, addrs[0]).Amount)
	assert.Equal(t, addrs[0], keeper.GetDeposit(ctx, proposalID, addrs[0]).Depositer)
	assert.Equal(t, fourSteak, keeper.GetProposal(ctx, proposalID).TotalDeposit)

	err = keeper.AddDeposit(ctx, proposalID, addrs[0], fiveSteak)
	assert.Nil(t, err)
	assert.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetDeposit(ctx, proposalID, addrs[0]).Amount)
	assert.Equal(t, addrs[0], keeper.GetDeposit(ctx, proposalID, addrs[0]).Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak), keeper.GetProposal(ctx, proposalID).TotalDeposit)

	err = keeper.AddDeposit(ctx, proposalID, addrs[1], fourSteak)
	assert.Nil(t, err)
	assert.Equal(t, fourSteak, keeper.GetDeposit(ctx, proposalID, addrs[1]).Amount)
	assert.Equal(t, addrs[1], keeper.GetDeposit(ctx, proposalID, addrs[1]).Depositer)
	assert.Equal(t, fourSteak.Plus(fiveSteak).Plus(fourSteak), keeper.GetProposal(ctx, proposalID).TotalDeposit)

	assert.Equal(t, ctx.BlockHeight(), keeper.GetProposal(ctx, proposalID).VotingStartBlock)
	assert.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	assert.Equal(t, proposalID, keeper.ActiveProposalQueuePeek(ctx).ProposalID)

}
