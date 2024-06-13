package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

const (
	baseDepositTestAmount  = 100
	baseDepositTestPercent = 25
)

func TestDeposits(t *testing.T) {
	govKeeper, _, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
	trackMockBalances(bankKeeper)
	TestAddrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(10000000))

	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "", "title", "summary", TestAddrs[0], false)
	require.NoError(t, err)
	proposalID := proposal.Id

	fourStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 4)))
	fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 5)))

	addr0Initial := bankKeeper.GetAllBalances(ctx, TestAddrs[0])
	addr1Initial := bankKeeper.GetAllBalances(ctx, TestAddrs[1])

	require.True(t, sdk.NewCoins(proposal.TotalDeposit...).IsEqual(sdk.NewCoins()))

	// Check no deposits at beginning
	deposit, found := govKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.False(t, found)
	proposal, ok := govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Nil(t, proposal.VotingStartTime)

	// Check first deposit
	votingStarted, err := govKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], fourStake)
	require.NoError(t, err)
	require.False(t, votingStarted)
	deposit, found = govKeeper.GetDeposit(ctx, proposalID, TestAddrs[0])
	require.True(t, found)
	require.Equal(t, fourStake, sdk.NewCoins(deposit.Amount...))
	require.Equal(t, TestAddrs[0].String(), deposit.Depositor)
	proposal, ok = govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake, sdk.NewCoins(proposal.TotalDeposit...))
	require.Equal(t, addr0Initial.Sub(fourStake...), bankKeeper.GetAllBalances(ctx, TestAddrs[0]))

	// Check a second deposit from same address
	votingStarted, err = govKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], fiveStake)
	require.NoError(t, err)
	require.False(t, votingStarted)
	deposit, found = govKeeper.GetDeposit(ctx, proposalID, TestAddrs[0])
	require.True(t, found)
	require.Equal(t, fourStake.Add(fiveStake...), sdk.NewCoins(deposit.Amount...))
	require.Equal(t, TestAddrs[0].String(), deposit.Depositor)
	proposal, ok = govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake...), sdk.NewCoins(proposal.TotalDeposit...))
	require.Equal(t, addr0Initial.Sub(fourStake...).Sub(fiveStake...), bankKeeper.GetAllBalances(ctx, TestAddrs[0]))

	// Check third deposit from a new address
	votingStarted, err = govKeeper.AddDeposit(ctx, proposalID, TestAddrs[1], fourStake)
	require.NoError(t, err)
	require.True(t, votingStarted)
	deposit, found = govKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.True(t, found)
	require.Equal(t, TestAddrs[1].String(), deposit.Depositor)
	require.Equal(t, fourStake, sdk.NewCoins(deposit.Amount...))
	proposal, ok = govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.Equal(t, fourStake.Add(fiveStake...).Add(fourStake...), sdk.NewCoins(proposal.TotalDeposit...))
	require.Equal(t, addr1Initial.Sub(fourStake...), bankKeeper.GetAllBalances(ctx, TestAddrs[1]))

	// Check that proposal moved to voting period
	proposal, ok = govKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	// Test deposit iterator
	// NOTE order of deposits is determined by the addresses
	deposits := govKeeper.GetAllDeposits(ctx)
	require.Len(t, deposits, 2)
	require.Equal(t, deposits, govKeeper.GetDeposits(ctx, proposalID))
	require.Equal(t, TestAddrs[0].String(), deposits[0].Depositor)
	require.Equal(t, fourStake.Add(fiveStake...), sdk.NewCoins(deposits[0].Amount...))
	require.Equal(t, TestAddrs[1].String(), deposits[1].Depositor)
	require.Equal(t, fourStake, sdk.NewCoins(deposits[1].Amount...))

	// Test Refund Deposits
	deposit, found = govKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.True(t, found)
	require.Equal(t, fourStake, sdk.NewCoins(deposit.Amount...))
	govKeeper.RefundAndDeleteDeposits(ctx, proposalID)
	deposit, found = govKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
	require.False(t, found)
	require.Equal(t, addr0Initial, bankKeeper.GetAllBalances(ctx, TestAddrs[0]))
	require.Equal(t, addr1Initial, bankKeeper.GetAllBalances(ctx, TestAddrs[1]))

	// Test delete and burn deposits
	proposal, err = govKeeper.SubmitProposal(ctx, tp, "", "title", "summary", TestAddrs[0], true)
	require.NoError(t, err)
	proposalID = proposal.Id
	_, err = govKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], fourStake)
	require.NoError(t, err)
	govKeeper.DeleteAndBurnDeposits(ctx, proposalID)
	deposits = govKeeper.GetDeposits(ctx, proposalID)
	require.Len(t, deposits, 0)
	require.Equal(t, addr0Initial.Sub(fourStake...), bankKeeper.GetAllBalances(ctx, TestAddrs[0]))
}

func TestDepositAmount(t *testing.T) {
	testcases := []struct {
		name            string
		deposit         sdk.Coins
		minDepositRatio string
		err             string
	}{
		{
			name:            "good amount and denoms",
			deposit:         sdk.NewCoins(sdk.NewInt64Coin("stake", 10000)),
			minDepositRatio: "0.001",
		},
		{
			name:            "good amount and denoms but not enough balance for zcoin",
			deposit:         sdk.NewCoins(sdk.NewInt64Coin("stake", 10000), sdk.NewInt64Coin("zcoin", 1)),
			minDepositRatio: "0.001",
			err:             "not enough balance",
		},
		{
			name:            "too small amount",
			deposit:         sdk.NewCoins(sdk.NewInt64Coin("stake", 10)),
			minDepositRatio: "0.001",
			err:             "received 10stake but need at least one of the following: 10000stake,10zcoin: minimum deposit is too small",
		},
		{
			name:            "too small amount with another coin",
			deposit:         sdk.NewCoins(sdk.NewInt64Coin("zcoin", 1)),
			minDepositRatio: "0.001",
			err:             "received 1zcoin but need at least one of the following: 10000stake,10zcoin: minimum deposit is too small",
		},
		{
			name:            "bad denom",
			deposit:         sdk.NewCoins(sdk.NewInt64Coin("euro", 10000)),
			minDepositRatio: "0.001",
			err:             "deposited 10000euro, but gov accepts only the following denom(s): [stake zcoin]: invalid deposit denom",
		},
		{
			name:            "mix containing bad and good denom",
			deposit:         sdk.NewCoins(sdk.NewInt64Coin("stake", 10000), sdk.NewInt64Coin("euro", 10000)),
			minDepositRatio: "0.001",
			err:             "deposited 10000euro,10000stake, but gov accepts only the following denom(s): [stake zcoin]: invalid deposit denom",
		},
		{
			name:            "minDepositRatio is zero",
			deposit:         sdk.NewCoins(sdk.NewInt64Coin("stake", 10)),
			minDepositRatio: "0.0",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			govKeeper, _, bankKeeper, stakingKeeper, _, ctx := setupGovKeeper(t)
			trackMockBalances(bankKeeper)

			testAddrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000000000000000))

			params := v1.DefaultParams()
			params.MinDepositRatio = tc.minDepositRatio
			params.MinDeposit = sdk.NewCoins(params.MinDeposit...).Add(sdk.NewCoin("zcoin", sdk.NewInt(10000))) // coins must be sorted by denom
			err := govKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			tp := TestProposal
			proposal, err := govKeeper.SubmitProposal(ctx, tp, "", "title", "summary", testAddrs[0], false)
			require.NoError(t, err)
			proposalID := proposal.Id

			_, err = govKeeper.AddDeposit(ctx, proposalID, testAddrs[0], tc.deposit)
			if tc.err != "" {
				require.Error(t, err)
				require.Equal(t, tc.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateInitialDeposit(t *testing.T) {
	testcases := map[string]struct {
		minDeposit               sdk.Coins
		minInitialDepositPercent int64
		initialDeposit           sdk.Coins
		expedited                bool

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
			govKeeper, _, _, _, _, ctx := setupGovKeeper(t)

			params := v1.DefaultParams()
			params.MinDeposit = tc.minDeposit
			params.MinInitialDepositRatio = sdk.NewDec(tc.minInitialDepositPercent).Quo(sdk.NewDec(100)).String()

			govKeeper.SetParams(ctx, params)

			err := govKeeper.ValidateInitialDeposit(ctx, tc.initialDeposit, tc.expedited)

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
