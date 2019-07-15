package keeper

import (
	"time"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDeposits(t *testing.T) {
	ctx, ak, keeper, _ := createTestInput(t, false, 100)

	tp := TestProposal()
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	fourStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(4)))
	fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5)))

	addr0Initial := ak.GetAccount(ctx, TestAddrs[0]).GetCoins()
	addr1Initial := ak.GetAccount(ctx, TestAddrs[1]).GetCoins()

	expTokens := sdk.TokensFromConsensusPower(42)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, expTokens)), addr0Initial)
	require.True(t, proposal.TotalDeposit.IsEqual(sdk.NewCoins()))

	// Check no deposits at beginning
	deposit, found := keeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.False(t, found)
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	// Check first deposit
	err, votingStarted := keeper.AddDeposit(ctx, proposalID, TestAddrs[0], fourStake)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, TestAddrs[0])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	require.Equal(t, TestAddrs[0], deposit.Depositor)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake, proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourStake), ak.GetAccount(ctx, TestAddrs[0]).GetCoins())

	// Check a second deposit from same address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, TestAddrs[0], fiveStake)
	require.Nil(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, TestAddrs[0])
	require.True(t, found)
	require.Equal(t, fourStake.Add(fiveStake), deposit.Amount)
	require.Equal(t, TestAddrs[0], deposit.Depositor)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake), proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourStake).Sub(fiveStake), ak.GetAccount(ctx, TestAddrs[0]).GetCoins())

	// Check third deposit from a new address
	err, votingStarted = keeper.AddDeposit(ctx, proposalID, TestAddrs[1], fourStake)
	require.Nil(t, err)
	require.True(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.True(t, found)
	require.Equal(t, TestAddrs[1], deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake).Add(fourStake), proposal.TotalDeposit)
	require.Equal(t, addr1Initial.Sub(fourStake), ak.GetAccount(ctx, TestAddrs[1]).GetCoins())

	// Check that proposal moved to voting period
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	// Test deposit iterator
	depositsIterator := keeper.GetDepositsIterator(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, TestAddrs[0], deposit.Depositor)
	require.Equal(t, fourStake.Add(fiveStake), deposit.Amount)
	depositsIterator.Next()
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
	require.Equal(t, TestAddrs[1], deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	depositsIterator.Next()
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()

	// Test Refund Deposits
	deposit, found = keeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	keeper.RefundDeposits(ctx, proposalID)
	deposit, found = keeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.False(t, found)
	require.Equal(t, addr0Initial, ak.GetAccount(ctx, TestAddrs[0]).GetCoins())
	require.Equal(t, addr1Initial, ak.GetAccount(ctx, TestAddrs[1]).GetCoins())

}