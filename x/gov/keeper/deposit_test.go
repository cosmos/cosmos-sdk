package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDeposits(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	TestAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId

	fourStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(4)))
	fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5)))

	addr0Initial := app.BankKeeper.GetAllBalances(ctx, TestAddrs[0])
	addr1Initial := app.BankKeeper.GetAllBalances(ctx, TestAddrs[1])

	require.True(t, proposal.TotalDeposit.IsEqual(sdk.NewCoins()))

	// Check no deposits at beginning
	deposit, found := app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.False(t, found)
	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	// Check first deposit
	votingStarted, err := app.GovKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], fourStake)
	require.NoError(t, err)
	require.False(t, votingStarted)
	deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[0])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	require.Equal(t, TestAddrs[0].String(), deposit.Depositor)
	proposal, ok = app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake, proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourStake), app.BankKeeper.GetAllBalances(ctx, TestAddrs[0]))

	// Check a second deposit from same address
	votingStarted, err = app.GovKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], fiveStake)
	require.NoError(t, err)
	require.False(t, votingStarted)
	deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[0])
	require.True(t, found)
	require.Equal(t, fourStake.Add(fiveStake...), deposit.Amount)
	require.Equal(t, TestAddrs[0].String(), deposit.Depositor)
	proposal, ok = app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake...), proposal.TotalDeposit)
	require.Equal(t, addr0Initial.Sub(fourStake).Sub(fiveStake), app.BankKeeper.GetAllBalances(ctx, TestAddrs[0]))

	// Check third deposit from a new address
	votingStarted, err = app.GovKeeper.AddDeposit(ctx, proposalID, TestAddrs[1], fourStake)
	require.NoError(t, err)
	require.True(t, votingStarted)
	deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.True(t, found)
	require.Equal(t, TestAddrs[1].String(), deposit.Depositor)
	require.Equal(t, fourStake, deposit.Amount)
	proposal, ok = app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake...).Add(fourStake...), proposal.TotalDeposit)
	require.Equal(t, addr1Initial.Sub(fourStake), app.BankKeeper.GetAllBalances(ctx, TestAddrs[1]))

	// Check that proposal moved to voting period
	proposal, ok = app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	// Test deposit iterator
	// NOTE order of deposits is determined by the addresses
	deposits := app.GovKeeper.GetAllDeposits(ctx)
	require.Len(t, deposits, 2)
	require.Equal(t, deposits, app.GovKeeper.GetDeposits(ctx, proposalID))
	require.Equal(t, TestAddrs[0].String(), deposits[0].Depositor)
	require.Equal(t, fourStake.Add(fiveStake...), deposits[0].Amount)
	require.Equal(t, TestAddrs[1].String(), deposits[1].Depositor)
	require.Equal(t, fourStake, deposits[1].Amount)

	// Test Refund Deposits
	deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.True(t, found)
	require.Equal(t, fourStake, deposit.Amount)
	app.GovKeeper.RefundDeposits(ctx, proposalID)
	deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.False(t, found)
	require.Equal(t, addr0Initial, app.BankKeeper.GetAllBalances(ctx, TestAddrs[0]))
	require.Equal(t, addr1Initial, app.BankKeeper.GetAllBalances(ctx, TestAddrs[1]))
}
