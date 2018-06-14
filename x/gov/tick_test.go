package gov

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)
	// stakeHandler := stake.NewHandler(keeper.sk)
	govHandler := NewHandler(keeper)

	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", "Text", addrs[0], sdk.Coins{sdk.Coin{"steak", 5}})

	res := govHandler(ctx, newProposalMsg)
	assert.True(t, res.IsOK())

	EndBlocker(ctx, keeper)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	ctx = ctx.WithBlockHeight(10)
	EndBlocker(ctx, keeper)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	ctx = ctx.WithBlockHeight(250)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)
	// stakeHandler := stake.NewHandler(keeper.sk)
	govHandler := NewHandler(keeper)

	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", "Text", addrs[0], sdk.Coins{sdk.Coin{"steak", 5}})

	res := govHandler(ctx, newProposalMsg)
	assert.True(t, res.IsOK())

	// var proposalID int
	// keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	EndBlocker(ctx, keeper)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	ctx = ctx.WithBlockHeight(10)
	EndBlocker(ctx, keeper)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg2 := NewMsgSubmitProposal("Test2", "test2", "Text", addrs[1], sdk.Coins{sdk.Coin{"steak", 5}})
	res = govHandler(ctx, newProposalMsg2)
	assert.True(t, res.IsOK())

	ctx = ctx.WithBlockHeight(205)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	ctx = ctx.WithBlockHeight(215)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickPassedDepositPeriod(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)
	// stakeHandler := stake.NewHandler(keeper.sk)
	govHandler := NewHandler(keeper)

	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", "Text", addrs[0], sdk.Coins{sdk.Coin{"steak", 5}})

	res := govHandler(ctx, newProposalMsg)
	assert.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	EndBlocker(ctx, keeper)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	ctx = ctx.WithBlockHeight(10)
	EndBlocker(ctx, keeper)
	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.Coin{"steak", 5}})
	res = govHandler(ctx, newDepositMsg)
	assert.True(t, res.IsOK())

	assert.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	assert.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))

	EndBlocker(ctx, keeper)

	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	assert.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopActiveProposalQueue(ctx, keeper))
}

func TestTickPassedVotingPeriod(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 100)
	// stakeHandler := stake.NewHandler(keeper.sk)
	govHandler := NewHandler(keeper)

	assert.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	assert.False(t, shouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", "Text", addrs[0], sdk.Coins{sdk.Coin{"steak", 5}})

	res := govHandler(ctx, newProposalMsg)
	assert.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	ctx = ctx.WithBlockHeight(10)
	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.Coin{"steak", 5}})
	res = govHandler(ctx, newDepositMsg)
	assert.True(t, res.IsOK())

	EndBlocker(ctx, keeper)

	ctx = ctx.WithBlockHeight(215)
	assert.True(t, shouldPopActiveProposalQueue(ctx, keeper))
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	assert.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	assert.Equal(t, StatusVotingPeriod, keeper.GetProposal(ctx, proposalID).Status)

	EndBlocker(ctx, keeper)

	assert.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	depositsIterator = keeper.GetDeposits(ctx, proposalID)
	assert.False(t, depositsIterator.Valid())
	depositsIterator.Close()
	assert.Equal(t, StatusRejected, keeper.GetProposal(ctx, proposalID).Status)
}
