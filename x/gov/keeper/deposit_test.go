package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDeposits(t *testing.T) {
	ctx, ak, keeper, _, _ := createTestInput(t, false, 100)

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	fourStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(4)))
	fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5)))

	addr0Initial := ak.GetAccount(ctx, TestAddrs[0]).GetCoins()
	addr1Initial := ak.GetAccount(ctx, TestAddrs[1]).GetCoins()

	require.True(t, proposal.TotalDeposit.IsEqual(sdk.NewCoins()))

	// Check no deposits at beginning
	deposit, found := keeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.False(t, found)
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	// Check first deposit
	votingStarted, err := keeper.AddDeposit(ctx, proposalID, TestAddrs[0], fourStake)
	require.NoError(t, err)
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
	votingStarted, err = keeper.AddDeposit(ctx, proposalID, TestAddrs[0], fiveStake)
	require.NoError(t, err)
	require.False(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, TestAddrs[0])
	require.True(t, found)
	require.Equal(t, fourStake.Add(fiveStake...), deposit.Amount)
	require.Equal(t, TestAddrs[0], deposit.Depositor)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake...), proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourStake).Sub(fiveStake), ak.GetAccount(ctx, TestAddrs[0]).GetCoins())

	// Check third deposit from a new address
	votingStarted, err = keeper.AddDeposit(ctx, proposalID, TestAddrs[1], fourStake)
	require.NoError(t, err)
	require.True(t, votingStarted)
	deposit, found = keeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.True(t, found)
	require.Equal(t, TestAddrs[1], deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake...).Add(fourStake...), proposal.TotalDeposit)
	require.Equal(t, addr1Initial.Sub(fourStake), ak.GetAccount(ctx, TestAddrs[1]).GetCoins())

	// Check that proposal moved to voting period
	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	// Test deposit iterator
	// NOTE order of deposits is determined by the addresses
	deposits := keeper.GetAllDeposits(ctx)
	require.Len(t, deposits, 2)
	require.Equal(t, deposits, keeper.GetDeposits(ctx, proposalID))
	require.Equal(t, TestAddrs[0], deposits[0].Depositor)
	require.Equal(t, fourStake.Add(fiveStake...), deposits[0].Amount)
	require.Equal(t, TestAddrs[1], deposits[1].Depositor)
	require.Equal(t, fourStake, deposits[1].Amount)

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
