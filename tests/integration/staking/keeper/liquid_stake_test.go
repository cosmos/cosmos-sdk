package keeper_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Helper function to clear the Bonded pool balances before a unit test
func clearPoolBalance(t *testing.T, sk keeper.Keeper, ak accountkeeper.AccountKeeper, bk bankkeeper.Keeper, ctx sdk.Context) {
	bondDenom, err := sk.BondDenom(ctx)
	require.NoError(t, err)
	initialBondedBalance := bk.GetBalance(ctx, ak.GetModuleAddress(types.BondedPoolName), bondDenom)

	err = bk.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, minttypes.ModuleName, sdk.NewCoins(initialBondedBalance))
	require.NoError(t, err, "no error expected when clearing bonded pool balance")
}

// Helper function to fund the Bonded pool balances before a unit test
func fundPoolBalance(t *testing.T, sk keeper.Keeper, bk bankkeeper.Keeper, ctx sdk.Context, amount math.Int) {
	bondDenom, err := sk.BondDenom(ctx)
	require.NoError(t, err)
	bondedPoolCoin := sdk.NewCoin(bondDenom, amount)

	err = bk.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(bondedPoolCoin))
	require.NoError(t, err, "no error expected when minting")

	err = bk.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.BondedPoolName, sdk.NewCoins(bondedPoolCoin))
	require.NoError(t, err, "no error expected when sending tokens to bonded pool")
}

// Helper function to create a module account address from a tokenized share
// Used to mock the delegation owner of a tokenized share
func createTokenizeShareModuleAccount(recordID uint64) sdk.AccAddress {
	record := types.TokenizeShareRecord{
		Id:            recordID,
		ModuleAccount: fmt.Sprintf("%s%d", types.TokenizeShareModuleAccountPrefix, recordID),
	}
	return record.GetModuleAddress()
}

// Helper function to create a base account from an account name
// Used to differentiate against liquid staking provider module account
func createBaseAccount(ak accountkeeper.AccountKeeper, ctx sdk.Context, accountName string) sdk.AccAddress {
	baseAccountAddress := sdk.AccAddress(accountName)
	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, baseAccountAddress))
	return baseAccountAddress
}

// Tests CheckExceedsGlobalLiquidStakingCap
func TestCheckExceedsGlobalLiquidStakingCap(t *testing.T) {
	var (
		accountKeeper accountkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *keeper.Keeper
	)

	app, err := simtestutil.Setup(
		depinject.Configs(testutil.AppConfig, depinject.Supply(log.NewNopLogger())),
		&accountKeeper,
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)
	ctx := app.BaseApp.NewContext(false)

	testCases := []struct {
		name             string
		globalLiquidCap  math.LegacyDec
		totalLiquidStake math.Int
		totalStake       math.Int
		newLiquidStake   math.Int
		tokenizingShares bool
		expectedExceeds  bool
	}{
		{
			// Cap: 10% - Native Delegation - Delegation Below Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 1
			// => Total Liquid Stake: 5+1=6, Total Stake: 95+1=96 => 6/96 = 6% < 10% cap
			name:             "10 percent cap _ native delegation _ delegation below cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.1"),
			totalLiquidStake: math.NewInt(5),
			totalStake:       math.NewInt(95),
			newLiquidStake:   math.NewInt(1),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 10% - Native Delegation - Delegation At Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 5
			// => Total Liquid Stake: 5+5=10, Total Stake: 95+5=100 => 10/100 = 10% == 10% cap
			name:             "10 percent cap _ native delegation _ delegation equals cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.1"),
			totalLiquidStake: math.NewInt(5),
			totalStake:       math.NewInt(95),
			newLiquidStake:   math.NewInt(5),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 10% - Native Delegation - Delegation Exceeds Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 6
			// => Total Liquid Stake: 5+6=11, Total Stake: 95+6=101 => 11/101 = 11% > 10% cap
			name:             "10 percent cap _ native delegation _ delegation exceeds cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.1"),
			totalLiquidStake: math.NewInt(5),
			totalStake:       math.NewInt(95),
			newLiquidStake:   math.NewInt(6),
			tokenizingShares: false,
			expectedExceeds:  true,
		},
		{
			// Cap: 20% - Native Delegation - Delegation Below Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 29
			// => Total Liquid Stake: 20+29=49, Total Stake: 220+29=249 => 49/249 = 19% < 20% cap
			name:             "20 percent cap _ native delegation _ delegation below cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.20"),
			totalLiquidStake: math.NewInt(20),
			totalStake:       math.NewInt(220),
			newLiquidStake:   math.NewInt(29),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 20% - Native Delegation - Delegation At Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 30
			// => Total Liquid Stake: 20+30=50, Total Stake: 220+30=250 => 50/250 = 20% == 20% cap
			name:             "20 percent cap _ native delegation _ delegation equals cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.20"),
			totalLiquidStake: math.NewInt(20),
			totalStake:       math.NewInt(220),
			newLiquidStake:   math.NewInt(30),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 20% - Native Delegation - Delegation Exceeds Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 31
			// => Total Liquid Stake: 20+31=51, Total Stake: 220+31=251 => 51/251 = 21% > 20% cap
			name:             "20 percent cap _ native delegation _ delegation exceeds cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.20"),
			totalLiquidStake: math.NewInt(20),
			totalStake:       math.NewInt(220),
			newLiquidStake:   math.NewInt(31),
			tokenizingShares: false,
			expectedExceeds:  true,
		},
		{
			// Cap: 50% - Native Delegation - Delegation Below Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 50
			// => Total Liquid Stake: 0+50=50, Total Stake: 100+50=150 => 50/150 = 33% < 50% cap
			name:             "50 percent cap _ native delegation _ delegation below cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.5"),
			totalLiquidStake: math.NewInt(0),
			totalStake:       math.NewInt(100),
			newLiquidStake:   math.NewInt(50),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 50% - Tokenized Delegation - Delegation At Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 50
			// => 50 / 100 = 50% == 50% cap
			name:             "50 percent cap _ tokenized delegation _ delegation equals cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.5"),
			totalLiquidStake: math.NewInt(0),
			totalStake:       math.NewInt(100),
			newLiquidStake:   math.NewInt(50),
			tokenizingShares: true,
			expectedExceeds:  false,
		},
		{
			// Cap: 50% - Native Delegation - Delegation Below Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 51
			// => Total Liquid Stake: 0+51=51, Total Stake: 100+51=151 => 51/151 = 33% < 50% cap
			name:             "50 percent cap _ native delegation _ delegation below cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.5"),
			totalLiquidStake: math.NewInt(0),
			totalStake:       math.NewInt(100),
			newLiquidStake:   math.NewInt(51),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 50% - Tokenized Delegation - Delegation Exceeds Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 51
			// => 51 / 100 = 51% > 50% cap
			name:             "50 percent cap _  tokenized delegation _delegation exceeds cap",
			globalLiquidCap:  math.LegacyMustNewDecFromStr("0.5"),
			totalLiquidStake: math.NewInt(0),
			totalStake:       math.NewInt(100),
			newLiquidStake:   math.NewInt(51),
			tokenizingShares: true,
			expectedExceeds:  true,
		},
		{
			// Cap of 0% - everything should exceed
			name:             "0 percent cap",
			globalLiquidCap:  math.LegacyZeroDec(),
			totalLiquidStake: math.NewInt(0),
			totalStake:       math.NewInt(1_000_000),
			newLiquidStake:   math.NewInt(1),
			tokenizingShares: false,
			expectedExceeds:  true,
		},
		{
			// Cap of 100% - nothing should exceed
			name:             "100 percent cap",
			globalLiquidCap:  math.LegacyOneDec(),
			totalLiquidStake: math.NewInt(1),
			totalStake:       math.NewInt(1),
			newLiquidStake:   math.NewInt(1_000_000),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the global liquid staking cap
			params, err := stakingKeeper.GetParams(ctx)
			require.NoError(t, err)
			params.GlobalLiquidStakingCap = tc.globalLiquidCap
			stakingKeeper.SetParams(ctx, params)

			// Update the total liquid tokens
			stakingKeeper.SetTotalLiquidStakedTokens(ctx, tc.totalLiquidStake)

			// Fund each pool for the given test case
			clearPoolBalance(t, *stakingKeeper, accountKeeper, bankKeeper, ctx)
			fundPoolBalance(t, *stakingKeeper, bankKeeper, ctx, tc.totalStake)

			// Check if the new tokens would exceed the global cap
			actualExceeds, err := stakingKeeper.CheckExceedsGlobalLiquidStakingCap(ctx, tc.newLiquidStake, tc.tokenizingShares)
			require.NoError(t, err)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests SafelyIncreaseTotalLiquidStakedTokens
func TestSafelyIncreaseTotalLiquidStakedTokens(t *testing.T) {
	var (
		accountKeeper accountkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *keeper.Keeper
	)

	app, err := simtestutil.Setup(
		depinject.Configs(testutil.AppConfig, depinject.Supply(log.NewNopLogger())),
		&accountKeeper,
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)
	ctx := app.BaseApp.NewContext(false)

	intitialTotalLiquidStaked := math.NewInt(100)
	increaseAmount := math.NewInt(10)
	poolBalance := math.NewInt(200)

	// Set the total staked and total liquid staked amounts
	// which are required components when checking the global cap
	// Total stake is calculated from the pool balance
	clearPoolBalance(t, *stakingKeeper, accountKeeper, bankKeeper, ctx)
	fundPoolBalance(t, *stakingKeeper, bankKeeper, ctx, poolBalance)
	stakingKeeper.SetTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)

	// Set the global cap such that a small delegation would exceed the cap
	params, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.GlobalLiquidStakingCap = math.LegacyMustNewDecFromStr("0.0001")
	stakingKeeper.SetParams(ctx, params)

	// Attempt to increase the total liquid stake again, it should error since
	// the cap was exceeded
	err = stakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount, true)
	require.ErrorIs(t, err, types.ErrGlobalLiquidStakingCapExceeded)
	require.Equal(t, intitialTotalLiquidStaked, stakingKeeper.GetTotalLiquidStakedTokens(ctx))

	// Now relax the cap so that the increase succeeds
	params.GlobalLiquidStakingCap = math.LegacyMustNewDecFromStr("0.99")
	stakingKeeper.SetParams(ctx, params)

	// Confirm the total increased
	err = stakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount, true)
	require.NoError(t, err)
	require.Equal(t, intitialTotalLiquidStaked.Add(increaseAmount), stakingKeeper.GetTotalLiquidStakedTokens(ctx))
}

// Test RefreshTotalLiquidStaked
func TestRefreshTotalLiquidStaked(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	ctx := f.sdkCtx

	var (
		accountKeeper = f.accountKeeper
		stakingKeeper = f.stakingKeeper
	)

	// Set an arbitrary total liquid staked tokens amount that will get overwritten by the refresh
	stakingKeeper.SetTotalLiquidStakedTokens(ctx, math.NewInt(999))

	// Add validator's with various exchange rates
	validators := []types.Validator{
		{
			// Exchange rate of 1
			OperatorAddress: "valA",
			Tokens:          math.NewInt(100),
			DelegatorShares: math.LegacyNewDec(100),
			LiquidShares:    math.LegacyNewDec(100), // should be overwritten
		},
		{
			// Exchange rate of 0.9
			OperatorAddress: "valB",
			Tokens:          math.NewInt(90),
			DelegatorShares: math.LegacyNewDec(100),
			LiquidShares:    math.LegacyNewDec(200), // should be overwritten
		},
		{
			// Exchange rate of 0.75
			OperatorAddress: "valC",
			Tokens:          math.NewInt(75),
			DelegatorShares: math.LegacyNewDec(100),
			LiquidShares:    math.LegacyNewDec(300), // should be overwritten
		},
	}

	// Add various delegations across the above validator's
	// Total Liquid Staked: 1,849 + 922 = 2,771
	// Liquid Shares:
	//   ValA: 400 + 325 = 725
	//   ValB: 860 + 580 = 1,440
	//   ValC: 900 + 100 = 1,000
	expectedTotalLiquidStaked := int64(2771)
	expectedValidatorLiquidShares := map[string]math.LegacyDec{
		"valA": math.LegacyNewDec(725),
		"valB": math.LegacyNewDec(1440),
		"valC": math.LegacyNewDec(1000),
	}

	delegations := []struct {
		delegation  types.Delegation
		isLSTP      bool
		isTokenized bool
	}{
		// Delegator A - Not a liquid staking provider
		// Number of tokens/shares is irrelevant for this test
		{
			isLSTP: false,
			delegation: types.Delegation{
				DelegatorAddress: "delA",
				ValidatorAddress: "valA",
				Shares:           math.LegacyNewDec(100),
			},
		},
		{
			isLSTP: false,
			delegation: types.Delegation{
				DelegatorAddress: "delA",
				ValidatorAddress: "valB",
				Shares:           math.LegacyNewDec(860),
			},
		},
		{
			isLSTP: false,
			delegation: types.Delegation{
				DelegatorAddress: "delA",
				ValidatorAddress: "valC",
				Shares:           math.LegacyNewDec(750),
			},
		},
		// Delegator B - Liquid staking provider, tokens included in total
		// Total liquid staked: 400 + 774 + 675 = 1,849
		{
			// Shares: 400 shares, Exchange Rate: 1.0, Tokens: 400
			isLSTP: true,
			delegation: types.Delegation{
				DelegatorAddress: "delB-LSTP",
				ValidatorAddress: "valA",
				Shares:           math.LegacyNewDec(400),
			},
		},
		{
			// Shares: 860 shares, Exchange Rate: 0.9, Tokens: 774
			isLSTP: true,
			delegation: types.Delegation{
				DelegatorAddress: "delB-LSTP",
				ValidatorAddress: "valB",
				Shares:           math.LegacyNewDec(860),
			},
		},
		{
			// Shares: 900 shares, Exchange Rate: 0.75, Tokens: 675
			isLSTP: true,
			delegation: types.Delegation{
				DelegatorAddress: "delB-LSTP",
				ValidatorAddress: "valC",
				Shares:           math.LegacyNewDec(900),
			},
		},
		// Delegator C - Tokenized shares, tokens included in total
		// Total liquid staked: 325 + 522 + 75 = 922
		{
			// Shares: 325 shares, Exchange Rate: 1.0, Tokens: 325
			isTokenized: true,
			delegation: types.Delegation{
				DelegatorAddress: "delC-LSTP",
				ValidatorAddress: "valA",
				Shares:           math.LegacyNewDec(325),
			},
		},
		{
			// Shares: 580 shares, Exchange Rate: 0.9, Tokens: 522
			isTokenized: true,
			delegation: types.Delegation{
				DelegatorAddress: "delC-LSTP",
				ValidatorAddress: "valB",
				Shares:           math.LegacyNewDec(580),
			},
		},
		{
			// Shares: 100 shares, Exchange Rate: 0.75, Tokens: 75
			isTokenized: true,
			delegation: types.Delegation{
				DelegatorAddress: "delC-LSTP",
				ValidatorAddress: "valC",
				Shares:           math.LegacyNewDec(100),
			},
		},
	}

	// Create validators based on the above (must use an actual validator address)
	addresses := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 5, f.stakingKeeper.TokensFromConsensusPower(ctx, 300))
	validatorAddresses := map[string]sdk.ValAddress{
		"valA": sdk.ValAddress(addresses[0]),
		"valB": sdk.ValAddress(addresses[1]),
		"valC": sdk.ValAddress(addresses[2]),
	}
	for _, validator := range validators {
		validator.OperatorAddress = validatorAddresses[validator.OperatorAddress].String()
		f.stakingKeeper.SetValidator(ctx, validator)
	}

	// Create the delegations based on the above (must use actual delegator addresses)
	for _, delegationCase := range delegations {
		var delegatorAddress sdk.AccAddress
		switch {
		case delegationCase.isLSTP:
			delegatorAddress = createICAAccount(ctx, accountKeeper)
		case delegationCase.isTokenized:
			delegatorAddress = createTokenizeShareModuleAccount(1)
		default:
			delegatorAddress = createBaseAccount(accountKeeper, ctx, delegationCase.delegation.DelegatorAddress)
		}

		delegation := delegationCase.delegation
		delegation.DelegatorAddress = delegatorAddress.String()
		delegation.ValidatorAddress = validatorAddresses[delegation.ValidatorAddress].String()
		f.stakingKeeper.SetDelegation(ctx, delegation)
	}

	// Refresh the total liquid staked and validator liquid shares
	err := f.stakingKeeper.RefreshTotalLiquidStaked(ctx)
	require.NoError(t, err, "no error expected when refreshing total liquid staked")

	// Check the total liquid staked and liquid shares by validator
	actualTotalLiquidStaked := f.stakingKeeper.GetTotalLiquidStakedTokens(ctx)
	require.Equal(t, expectedTotalLiquidStaked, actualTotalLiquidStaked.Int64(), "total liquid staked tokens")

	for _, moniker := range []string{"valA", "valB", "valC"} {
		address := validatorAddresses[moniker]
		expectedLiquidShares := expectedValidatorLiquidShares[moniker]

		actualValidator, err := f.stakingKeeper.GetValidator(ctx, address)
		require.True(t, !errors.Is(err, types.ErrNoValidatorFound), "validator %s should have been found after refresh", moniker)

		actualLiquidShares := actualValidator.LiquidShares
		require.Equal(t, expectedLiquidShares.TruncateInt64(), actualLiquidShares.TruncateInt64(),
			"liquid staked shares for validator %s", moniker)
	}
}
