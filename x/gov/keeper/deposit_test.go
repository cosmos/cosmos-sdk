package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

const (
	baseDepositTestAmount  = 100
	baseDepositTestPercent = 25
)

func TestDeposits(t *testing.T) {
	testcases := []struct {
		name      string
		expedited bool
	}{
		{
			name: "regular",
		},
		{
			name:      "expedited",
			expedited: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			govKeeper, authKeeper, bankKeeper, stakingKeeper, distKeeper, _, ctx := setupGovKeeper(t)
			trackMockBalances(bankKeeper, distKeeper)

			// With expedited proposals the minimum deposit is higher, so we must
			// initialize and deposit an amount depositMultiplier times larger
			// than the regular min deposit amount.
			depositMultiplier := int64(1)
			if tc.expedited {
				depositMultiplier = v1.DefaultMinExpeditedDepositTokensRatio
			}

			TestAddrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, sdkmath.NewInt(10000000*depositMultiplier))
			for _, addr := range TestAddrs {
				authKeeper.EXPECT().BytesToString(addr).Return(addr.String(), nil).AnyTimes()
				authKeeper.EXPECT().StringToBytes(addr.String()).Return(addr, nil).AnyTimes()
			}

			tp := TestProposal
			proposal, err := govKeeper.SubmitProposal(ctx, tp, "", "title", "summary", TestAddrs[0], tc.expedited)
			require.NoError(t, err)
			proposalID := proposal.Id

			fourStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 4*depositMultiplier)))
			fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 5*depositMultiplier)))

			addr0Initial := bankKeeper.GetAllBalances(ctx, TestAddrs[0])
			addr1Initial := bankKeeper.GetAllBalances(ctx, TestAddrs[1])

			require.True(t, sdk.NewCoins(proposal.TotalDeposit...).Equal(sdk.NewCoins()))

			// Check no deposits at beginning
			_, found := govKeeper.GetDeposit(ctx, proposalID, TestAddrs[1])
			require.False(t, found)
			proposal, ok := govKeeper.GetProposal(ctx, proposalID)
			require.True(t, ok)
			require.Nil(t, proposal.VotingStartTime)

			// Check first deposit
			votingStarted, err := govKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], fourStake)
			require.NoError(t, err)
			require.False(t, votingStarted)
			deposit, found := govKeeper.GetDeposit(ctx, proposalID, TestAddrs[0])
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
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),
		},
		"min deposit * initial percent < initial deposit: success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100+1))),
		},
		"min deposit * initial percent > initial deposit: error": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100-1))),

			expectError: true,
		},
		"min deposit * initial percent == initial deposit (non-base values and denom): success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin("uosmo", sdkmath.NewInt(56912))),
			minInitialDepositPercent: 50,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin("uosmo", sdkmath.NewInt(56912/2+10))),
		},
		"min deposit * initial percent == initial deposit but different denoms: error": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin("uosmo", sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),

			expectError: true,
		},
		"min deposit * initial percent == initial deposit (multiple coins): success": {
			minDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount)),
				sdk.NewCoin("uosmo", sdkmath.NewInt(baseDepositTestAmount*2))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100)),
				sdk.NewCoin("uosmo", sdkmath.NewInt(baseDepositTestAmount*2*baseDepositTestPercent/100)),
			),
		},
		"min deposit * initial percent > initial deposit (multiple coins): error": {
			minDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount)),
				sdk.NewCoin("uosmo", sdkmath.NewInt(baseDepositTestAmount*2))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100)),
				sdk.NewCoin("uosmo", sdkmath.NewInt(baseDepositTestAmount*2*baseDepositTestPercent/100-1)),
			),

			expectError: true,
		},
		"min deposit * initial percent < initial deposit (multiple coins - coin not required by min deposit): success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100)),
				sdk.NewCoin("uosmo", sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100-1)),
			),
		},
		"0 initial percent: success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: 0,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),
		},
		"expedited min deposit * initial percent == initial deposit: success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: baseDepositTestPercent,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),
			expedited:                true,
		},
		"expedited - 0 initial percent: success": {
			minDeposit:               sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositPercent: 0,
			initialDeposit:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount*baseDepositTestPercent/100))),
			expedited:                true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			govKeeper, _, _, _, _, _, ctx := setupGovKeeper(t)

			params := v1.DefaultParams()
			if tc.expedited {
				params.ExpeditedMinDeposit = tc.minDeposit
			} else {
				params.MinDeposit = tc.minDeposit
			}
			params.MinInitialDepositRatio = sdkmath.LegacyNewDec(tc.minInitialDepositPercent).Quo(sdkmath.LegacyNewDec(100)).String()

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

func TestChargeDeposit(t *testing.T) {
	testCases := []struct {
		name                      string
		proposalCancelRatio       string
		proposalCancelDestAddress string
		expectError               bool
	}{
		{
			name:                      "Success: CancelRatio=0",
			proposalCancelRatio:       "0",
			proposalCancelDestAddress: "",
			expectError:               false,
		},
		{
			name:                      "Success: CancelRatio=0.5",
			proposalCancelRatio:       "0.5",
			proposalCancelDestAddress: "",
			expectError:               false,
		},
		{
			name:                      "Success: CancelRatio=1",
			proposalCancelRatio:       "1",
			proposalCancelDestAddress: "",
			expectError:               false,
		},
	}

	for _, tc := range testCases {
		for i := 0; i < 3; i++ {
			testName := func(i int) string {
				if i == 0 {
					return fmt.Sprintf("%s and dest address is %s", tc.name, "nil")
				} else if i == 1 {
					return fmt.Sprintf("%s and dest address is normal address", tc.name)
				}
				return fmt.Sprintf("%s and dest address is community address", tc.name)
			}

			t.Run(testName(i), func(t *testing.T) {
				govKeeper, authKeeper, bankKeeper, stakingKeeper, _, _, ctx := setupGovKeeper(t)
				params := v1.DefaultParams()
				params.ProposalCancelRatio = tc.proposalCancelRatio
				TestAddrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(10000000000))
				for _, addr := range TestAddrs {
					authKeeper.EXPECT().BytesToString(addr).Return(addr.String(), nil).AnyTimes()
					authKeeper.EXPECT().StringToBytes(addr.String()).Return(addr, nil).AnyTimes()
				}

				switch i {
				case 0:
					// no dest address for cancel proposal, total cancellation charges will be burned
					params.ProposalCancelDest = ""
				case 1:
					// normal account address for proposal cancel dest address
					params.ProposalCancelDest = TestAddrs[1].String()
				default:
					// community address for proposal cancel dest address
					params.ProposalCancelDest = authtypes.NewModuleAddress(disttypes.ModuleName).String()
				}

				err := govKeeper.SetParams(ctx, params)
				require.NoError(t, err)

				tp := TestProposal
				proposal, err := govKeeper.SubmitProposal(ctx, tp, "", "title", "summary", TestAddrs[0], false)
				require.NoError(t, err)
				proposalID := proposal.Id
				// deposit to proposal
				fiveStake := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 300)))
				_, err = govKeeper.AddDeposit(ctx, proposalID, TestAddrs[0], fiveStake)
				require.NoError(t, err)

				codec := address.NewBech32Codec("cosmos")
				// get balances of dest address
				var prevBalance sdk.Coin
				if len(params.ProposalCancelDest) != 0 {
					accAddr, err := codec.StringToBytes(params.ProposalCancelDest)
					require.NoError(t, err)
					prevBalance = bankKeeper.GetBalance(ctx, accAddr, sdk.DefaultBondDenom)
				}

				// get the deposits
				allDeposits := govKeeper.GetDeposits(ctx, proposalID)

				// charge cancellation charges for cancel proposal
				err = govKeeper.ChargeDeposit(ctx, proposalID, TestAddrs[0].String(), params.ProposalCancelRatio)
				if tc.expectError {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				if len(params.ProposalCancelDest) != 0 {
					accAddr, err := codec.StringToBytes(params.ProposalCancelDest)
					require.NoError(t, err)
					newBalanceAfterCancelProposal := bankKeeper.GetBalance(ctx, accAddr, sdk.DefaultBondDenom)
					cancellationCharges := sdkmath.NewInt(0)
					for _, deposits := range allDeposits {
						for _, deposit := range deposits.Amount {
							burnAmount := sdkmath.LegacyNewDecFromInt(deposit.Amount).Mul(sdkmath.LegacyMustNewDecFromStr(params.MinInitialDepositRatio)).TruncateInt()
							cancellationCharges = cancellationCharges.Add(burnAmount)
						}
					}
					require.True(t, newBalanceAfterCancelProposal.Equal(prevBalance.Add(sdk.NewCoin(sdk.DefaultBondDenom, cancellationCharges))))
				}
			})
		}
	}
}
