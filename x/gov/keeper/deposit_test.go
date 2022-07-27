package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	baseDepositTestAmount  = 100
	baseDepositTestPercent = 25
)

func TestDeposits(t *testing.T) {
	testcases := map[string]struct {
		isExpedited bool
	}{
		"regular": {
			isExpedited: false,
		},
		"expedited": {
			isExpedited: true,
		},
	}

	for _, tc := range testcases {
		app := simapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})

		// With expedited proposals the minimum deposit is higer, so we must
		// initialize and deposit an amount depositMultiplier times larger
		// than the regular min deposit amount.
		depositMultiplier := int64(1)
		if tc.isExpedited {
			depositMultiplier = types.DefaultMinExpeditedDepositTokens.Quo(types.DefaultMinDepositTokens).Int64()
		}

		TestAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000000*depositMultiplier))

		tp := TestProposal
		proposal, err := app.GovKeeper.SubmitProposal(ctx, tp, tc.isExpedited)
		require.NoError(t, err)
		proposalID := proposal.ProposalId

		firstDepositValue := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, app.StakingKeeper.TokensFromConsensusPower(ctx, 4*depositMultiplier)))
		secondDepositValue := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, app.StakingKeeper.TokensFromConsensusPower(ctx, 5*depositMultiplier)))

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
		votingStarted, err := app.GovKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], firstDepositValue)
		require.NoError(t, err)
		require.False(t, votingStarted)
		deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[0])
		require.True(t, found)
		require.Equal(t, firstDepositValue, deposit.Amount)
		require.Equal(t, TestAddrs[0].String(), deposit.Depositor)
		proposal, ok = app.GovKeeper.GetProposal(ctx, proposalID)
		require.True(t, ok)
		require.Equal(t, firstDepositValue, proposal.TotalDeposit)
		require.Equal(t, addr0Initial.Sub(firstDepositValue), app.BankKeeper.GetAllBalances(ctx, TestAddrs[0]))

		// Check a second deposit from same address
		votingStarted, err = app.GovKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], secondDepositValue)
		require.NoError(t, err)
		require.False(t, votingStarted)
		deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[0])
		require.True(t, found)
		require.Equal(t, firstDepositValue.Add(secondDepositValue...), deposit.Amount)
		require.Equal(t, TestAddrs[0].String(), deposit.Depositor)
		proposal, ok = app.GovKeeper.GetProposal(ctx, proposalID)
		require.True(t, ok)
		require.Equal(t, firstDepositValue.Add(secondDepositValue...), proposal.TotalDeposit)
		require.Equal(t, addr0Initial.Sub(firstDepositValue).Sub(secondDepositValue), app.BankKeeper.GetAllBalances(ctx, TestAddrs[0]))

		// Check third deposit from a new address
		votingStarted, err = app.GovKeeper.AddDeposit(ctx, proposalID, TestAddrs[1], firstDepositValue)
		require.NoError(t, err)
		require.True(t, votingStarted)
		deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
		require.True(t, found)
		require.Equal(t, TestAddrs[1].String(), deposit.Depositor)
		require.Equal(t, firstDepositValue, deposit.Amount)
		proposal, ok = app.GovKeeper.GetProposal(ctx, proposalID)
		require.True(t, ok)
		require.Equal(t, firstDepositValue.Add(secondDepositValue...).Add(firstDepositValue...), proposal.TotalDeposit)
		require.Equal(t, addr1Initial.Sub(firstDepositValue), app.BankKeeper.GetAllBalances(ctx, TestAddrs[1]))

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
		require.Equal(t, firstDepositValue.Add(secondDepositValue...), deposits[0].Amount)
		require.Equal(t, TestAddrs[1].String(), deposits[1].Depositor)
		require.Equal(t, firstDepositValue, deposits[1].Amount)

		// Test Refund Deposits
		deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
		require.True(t, found)
		require.Equal(t, firstDepositValue, deposit.Amount)
		app.GovKeeper.RefundDeposits(ctx, proposalID)
		deposit, found = app.GovKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
		require.False(t, found)
		require.Equal(t, addr0Initial, app.BankKeeper.GetAllBalances(ctx, TestAddrs[0]))
		require.Equal(t, addr1Initial, app.BankKeeper.GetAllBalances(ctx, TestAddrs[1]))

		// Test delete deposits
		_, err = app.GovKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], firstDepositValue)
		require.NoError(t, err)
		app.GovKeeper.DeleteDeposits(ctx, proposalID)
		deposits = app.GovKeeper.GetDeposits(ctx, proposalID)
		require.Len(t, deposits, 0)
	}
}

func TestValidateInitialDeposit(t *testing.T) {
	testcases := map[string]struct {
		minDeposit               sdk.Coins
		minInitialDepositPercent int64
		initialDeposit           sdk.Coins

		expectError bool
	}{
		"min deposit * initial percent == initial deposit: success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),
		},
		"min deposit * initial percent < initial deposit: success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100+1))),
		},
		"min deposit * initial percent > initial deposit: error": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100-1))),

			expectError: true,
		},
		"min deposit * initial percent == initial deposit (non-base values and denom): success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(56912))),
			minInitialDepositPercent: 50,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(56912/2+10))),
		},
		"min deposit * initial percent == initial deposit but different denoms: error": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),

			expectError: true,
		},
		"min deposit * initial percent == initial deposit (multiple coins): success": {
			minDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount)),
				sdk.NewCoin("uosmo", sdk.NewInt(baseDepositTestAmount*2))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100)),
				sdk.NewCoin("uosmo", sdk.NewInt(baseDepositTestAmount*2*baseDepositTestPercent/100)),
			),
		},
		"min deposit * initial percent > initial deposit (multiple coins): error": {
			minDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount)),
				sdk.NewCoin("uosmo", sdk.NewInt(baseDepositTestAmount*2))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100)),
				sdk.NewCoin("uosmo", sdk.NewInt(baseDepositTestAmount*2*baseDepositTestPercent/100-1)),
			),

			expectError: true,
		},
		"min deposit * initial percent < initial deposit (multiple coins - coin not required by min deposit): success": {
			minDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100)),
				sdk.NewCoin("uosmo", sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100-1)),
			),
		},
		"0 initial percent: success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: 0,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			app := simapp.Setup(false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{})

			govKeeper := app.GovKeeper

			params := types.DefaultDepositParams()
			params.MinDeposit = tc.minDeposit
			params.MinInitialDepositRatio = sdk.NewDec(tc.minInitialDepositPercent).Quo(sdk.NewDec(100))

			govKeeper.SetDepositParams(ctx, params)

			err := govKeeper.ValidateInitialDeposit(ctx, tc.initialDeposit)

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
