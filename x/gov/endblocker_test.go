package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg2 := NewMsgSubmitProposal("Test2", "test2", ProposalTypeText, addrs[1], sdk.Coins{sdk.NewInt64Coin("steak", 5)})
	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickPassedDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, shouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin("steak", 5)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))

	EndBlocker(ctx, keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, shouldPopActiveProposalQueue(ctx, keeper))

}

func TestTickPassedVotingPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	SortAddresses(addrs)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, shouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin("steak", 5)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	EndBlocker(ctx, keeper)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod).Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.True(t, shouldPopActiveProposalQueue(ctx, keeper))
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, StatusVotingPeriod, keeper.GetProposal(ctx, proposalID).GetStatus())

	EndBlocker(ctx, keeper)

	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	depositsIterator = keeper.GetDeposits(ctx, proposalID)
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, StatusRejected, keeper.GetProposal(ctx, proposalID).GetStatus())
	require.True(t, keeper.GetProposal(ctx, proposalID).GetTallyResult().Equals(EmptyTallyResult()))
}
