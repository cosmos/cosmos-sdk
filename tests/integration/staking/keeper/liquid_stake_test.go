package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// Helper function to clear the Bonded pool balances before a unit test
func clearPoolBalance(t *testing.T, sk keeper.Keeper, ak accountkeeper.AccountKeeper, bk bankkeeper.Keeper, ctx sdk.Context) {
	bondDenom := sk.BondDenom(ctx)
	initialBondedBalance := bk.GetBalance(ctx, ak.GetModuleAddress(types.BondedPoolName), bondDenom)

	err := bk.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, minttypes.ModuleName, sdk.NewCoins(initialBondedBalance))
	require.NoError(t, err, "no error expected when clearing bonded pool balance")
}

// Helper function to fund the Bonded pool balances before a unit test
func fundPoolBalance(t *testing.T, sk keeper.Keeper, bk bankkeeper.Keeper, ctx sdk.Context, amount math.Int) {
	bondDenom := sk.BondDenom(ctx)
	bondedPoolCoin := sdk.NewCoin(bondDenom, amount)

	err := bk.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(bondedPoolCoin))
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
	ak.SetAccount(ctx, authtypes.NewBaseAccountWithAddress(baseAccountAddress))
	return baseAccountAddress
}

// Tests CheckExceedsGlobalLiquidStakingCap
func TestCheckExceedsGlobalLiquidStakingCap(t *testing.T) {
	var (
		accountKeeper accountkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *keeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	testCases := []struct {
		name             string
		globalLiquidCap  sdk.Dec
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
			globalLiquidCap:  sdk.MustNewDecFromStr("0.1"),
			totalLiquidStake: sdk.NewInt(5),
			totalStake:       sdk.NewInt(95),
			newLiquidStake:   sdk.NewInt(1),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 10% - Native Delegation - Delegation At Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 5
			// => Total Liquid Stake: 5+5=10, Total Stake: 95+5=100 => 10/100 = 10% == 10% cap
			name:             "10 percent cap _ native delegation _ delegation equals cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.1"),
			totalLiquidStake: sdk.NewInt(5),
			totalStake:       sdk.NewInt(95),
			newLiquidStake:   sdk.NewInt(5),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 10% - Native Delegation - Delegation Exceeds Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 6
			// => Total Liquid Stake: 5+6=11, Total Stake: 95+6=101 => 11/101 = 11% > 10% cap
			name:             "10 percent cap _ native delegation _ delegation exceeds cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.1"),
			totalLiquidStake: sdk.NewInt(5),
			totalStake:       sdk.NewInt(95),
			newLiquidStake:   sdk.NewInt(6),
			tokenizingShares: false,
			expectedExceeds:  true,
		},
		{
			// Cap: 20% - Native Delegation - Delegation Below Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 29
			// => Total Liquid Stake: 20+29=49, Total Stake: 220+29=249 => 49/249 = 19% < 20% cap
			name:             "20 percent cap _ native delegation _ delegation below cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.20"),
			totalLiquidStake: sdk.NewInt(20),
			totalStake:       sdk.NewInt(220),
			newLiquidStake:   sdk.NewInt(29),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 20% - Native Delegation - Delegation At Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 30
			// => Total Liquid Stake: 20+30=50, Total Stake: 220+30=250 => 50/250 = 20% == 20% cap
			name:             "20 percent cap _ native delegation _ delegation equals cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.20"),
			totalLiquidStake: sdk.NewInt(20),
			totalStake:       sdk.NewInt(220),
			newLiquidStake:   sdk.NewInt(30),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 20% - Native Delegation - Delegation Exceeds Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 31
			// => Total Liquid Stake: 20+31=51, Total Stake: 220+31=251 => 51/251 = 21% > 20% cap
			name:             "20 percent cap _ native delegation _ delegation exceeds cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.20"),
			totalLiquidStake: sdk.NewInt(20),
			totalStake:       sdk.NewInt(220),
			newLiquidStake:   sdk.NewInt(31),
			tokenizingShares: false,
			expectedExceeds:  true,
		},
		{
			// Cap: 50% - Native Delegation - Delegation Below Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 50
			// => Total Liquid Stake: 0+50=50, Total Stake: 100+50=150 => 50/150 = 33% < 50% cap
			name:             "50 percent cap _ native delegation _ delegation below cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.5"),
			totalLiquidStake: sdk.NewInt(0),
			totalStake:       sdk.NewInt(100),
			newLiquidStake:   sdk.NewInt(50),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 50% - Tokenized Delegation - Delegation At Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 50
			// => 50 / 100 = 50% == 50% cap
			name:             "50 percent cap _ tokenized delegation _ delegation equals cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.5"),
			totalLiquidStake: sdk.NewInt(0),
			totalStake:       sdk.NewInt(100),
			newLiquidStake:   sdk.NewInt(50),
			tokenizingShares: true,
			expectedExceeds:  false,
		},
		{
			// Cap: 50% - Native Delegation - Delegation Below Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 51
			// => Total Liquid Stake: 0+51=51, Total Stake: 100+51=151 => 51/151 = 33% < 50% cap
			name:             "50 percent cap _ native delegation _ delegation below cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.5"),
			totalLiquidStake: sdk.NewInt(0),
			totalStake:       sdk.NewInt(100),
			newLiquidStake:   sdk.NewInt(51),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
		{
			// Cap: 50% - Tokenized Delegation - Delegation Exceeds Threshold
			// Total Liquid Stake: 0, Total Stake: 100, New Liquid Stake: 51
			// => 51 / 100 = 51% > 50% cap
			name:             "50 percent cap _  tokenized delegation _delegation exceeds cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.5"),
			totalLiquidStake: sdk.NewInt(0),
			totalStake:       sdk.NewInt(100),
			newLiquidStake:   sdk.NewInt(51),
			tokenizingShares: true,
			expectedExceeds:  true,
		},
		{
			// Cap of 0% - everything should exceed
			name:             "0 percent cap",
			globalLiquidCap:  sdk.ZeroDec(),
			totalLiquidStake: sdk.NewInt(0),
			totalStake:       sdk.NewInt(1_000_000),
			newLiquidStake:   sdk.NewInt(1),
			tokenizingShares: false,
			expectedExceeds:  true,
		},
		{
			// Cap of 100% - nothing should exceed
			name:             "100 percent cap",
			globalLiquidCap:  sdk.OneDec(),
			totalLiquidStake: sdk.NewInt(1),
			totalStake:       sdk.NewInt(1),
			newLiquidStake:   sdk.NewInt(1_000_000),
			tokenizingShares: false,
			expectedExceeds:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the global liquid staking cap
			params := stakingKeeper.GetParams(ctx)
			params.GlobalLiquidStakingCap = tc.globalLiquidCap
			stakingKeeper.SetParams(ctx, params)

			// Update the total liquid tokens
			stakingKeeper.SetTotalLiquidStakedTokens(ctx, tc.totalLiquidStake)

			// Fund each pool for the given test case
			clearPoolBalance(t, *stakingKeeper, accountKeeper, bankKeeper, ctx)
			fundPoolBalance(t, *stakingKeeper, bankKeeper, ctx, tc.totalStake)

			// Check if the new tokens would exceed the global cap
			actualExceeds := stakingKeeper.CheckExceedsGlobalLiquidStakingCap(ctx, tc.newLiquidStake, tc.tokenizingShares)
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

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	intitialTotalLiquidStaked := sdk.NewInt(100)
	increaseAmount := sdk.NewInt(10)
	poolBalance := sdk.NewInt(200)

	// Set the total staked and total liquid staked amounts
	// which are required components when checking the global cap
	// Total stake is calculated from the pool balance
	clearPoolBalance(t, *stakingKeeper, accountKeeper, bankKeeper, ctx)
	fundPoolBalance(t, *stakingKeeper, bankKeeper, ctx, poolBalance)
	stakingKeeper.SetTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)

	// Set the global cap such that a small delegation would exceed the cap
	params := stakingKeeper.GetParams(ctx)
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.0001")
	stakingKeeper.SetParams(ctx, params)

	// Attempt to increase the total liquid stake again, it should error since
	// the cap was exceeded
	err = stakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount, true)
	require.ErrorIs(t, err, types.ErrGlobalLiquidStakingCapExceeded)
	require.Equal(t, intitialTotalLiquidStaked, stakingKeeper.GetTotalLiquidStakedTokens(ctx))

	// Now relax the cap so that the increase succeeds
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.99")
	stakingKeeper.SetParams(ctx, params)

	// Confirm the total increased
	err = stakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount, true)
	require.NoError(t, err)
	require.Equal(t, intitialTotalLiquidStaked.Add(increaseAmount), stakingKeeper.GetTotalLiquidStakedTokens(ctx))
}

// Test RefreshTotalLiquidStaked
func TestRefreshTotalLiquidStaked(t *testing.T) {
	_, app, ctx := createTestInput(t)

	var (
		accountKeeper = app.AccountKeeper
		stakingKeeper = app.StakingKeeper
	)

	// Set an arbitrary total liquid staked tokens amount that will get overwritten by the refresh
	stakingKeeper.SetTotalLiquidStakedTokens(ctx, sdk.NewInt(999))

	// Add validator's with various exchange rates
	validators := []types.Validator{
		{
			// Exchange rate of 1
			OperatorAddress: "valA",
			Tokens:          sdk.NewInt(100),
			DelegatorShares: sdk.NewDec(100),
			LiquidShares:    sdk.NewDec(100), // should be overwritten
		},
		{
			// Exchange rate of 0.9
			OperatorAddress: "valB",
			Tokens:          sdk.NewInt(90),
			DelegatorShares: sdk.NewDec(100),
			LiquidShares:    sdk.NewDec(200), // should be overwritten
		},
		{
			// Exchange rate of 0.75
			OperatorAddress: "valC",
			Tokens:          sdk.NewInt(75),
			DelegatorShares: sdk.NewDec(100),
			LiquidShares:    sdk.NewDec(300), // should be overwritten
		},
	}

	// Add various delegations across the above validator's
	// Total Liquid Staked: 1,849 + 922 = 2,771
	// Liquid Shares:
	//   ValA: 400 + 325 = 725
	//   ValB: 860 + 580 = 1,440
	//   ValC: 900 + 100 = 1,000
	expectedTotalLiquidStaked := int64(2771)
	expectedValidatorLiquidShares := map[string]sdk.Dec{
		"valA": sdk.NewDec(725),
		"valB": sdk.NewDec(1440),
		"valC": sdk.NewDec(1000),
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
				Shares:           sdk.NewDec(100),
			},
		},
		{
			isLSTP: false,
			delegation: types.Delegation{
				DelegatorAddress: "delA",
				ValidatorAddress: "valB",
				Shares:           sdk.NewDec(860),
			},
		},
		{
			isLSTP: false,
			delegation: types.Delegation{
				DelegatorAddress: "delA",
				ValidatorAddress: "valC",
				Shares:           sdk.NewDec(750),
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
				Shares:           sdk.NewDec(400),
			},
		},
		{
			// Shares: 860 shares, Exchange Rate: 0.9, Tokens: 774
			isLSTP: true,
			delegation: types.Delegation{
				DelegatorAddress: "delB-LSTP",
				ValidatorAddress: "valB",
				Shares:           sdk.NewDec(860),
			},
		},
		{
			// Shares: 900 shares, Exchange Rate: 0.75, Tokens: 675
			isLSTP: true,
			delegation: types.Delegation{
				DelegatorAddress: "delB-LSTP",
				ValidatorAddress: "valC",
				Shares:           sdk.NewDec(900),
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
				Shares:           sdk.NewDec(325),
			},
		},
		{
			// Shares: 580 shares, Exchange Rate: 0.9, Tokens: 522
			isTokenized: true,
			delegation: types.Delegation{
				DelegatorAddress: "delC-LSTP",
				ValidatorAddress: "valB",
				Shares:           sdk.NewDec(580),
			},
		},
		{
			// Shares: 100 shares, Exchange Rate: 0.75, Tokens: 75
			isTokenized: true,
			delegation: types.Delegation{
				DelegatorAddress: "delC-LSTP",
				ValidatorAddress: "valC",
				Shares:           sdk.NewDec(100),
			},
		},
	}

	// Create validators based on the above (must use an actual validator address)
	addresses := simapp.AddTestAddrsIncremental(app, ctx, 5, app.StakingKeeper.TokensFromConsensusPower(ctx, 300))
	validatorAddresses := map[string]sdk.ValAddress{
		"valA": sdk.ValAddress(addresses[0]),
		"valB": sdk.ValAddress(addresses[1]),
		"valC": sdk.ValAddress(addresses[2]),
	}
	for _, validator := range validators {
		validator.OperatorAddress = validatorAddresses[validator.OperatorAddress].String()
		app.StakingKeeper.SetValidator(ctx, validator)
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
		app.StakingKeeper.SetDelegation(ctx, delegation)
	}

	// Refresh the total liquid staked and validator liquid shares
	err := app.StakingKeeper.RefreshTotalLiquidStaked(ctx)
	require.NoError(t, err, "no error expected when refreshing total liquid staked")

	// Check the total liquid staked and liquid shares by validator
	actualTotalLiquidStaked := app.StakingKeeper.GetTotalLiquidStakedTokens(ctx)
	require.Equal(t, expectedTotalLiquidStaked, actualTotalLiquidStaked.Int64(), "total liquid staked tokens")

	for _, moniker := range []string{"valA", "valB", "valC"} {
		address := validatorAddresses[moniker]
		expectedLiquidShares := expectedValidatorLiquidShares[moniker]

		actualValidator, found := app.StakingKeeper.GetValidator(ctx, address)
		require.True(t, found, "validator %s should have been found after refresh", moniker)

		actualLiquidShares := actualValidator.LiquidShares
		require.Equal(t, expectedLiquidShares.TruncateInt64(), actualLiquidShares.TruncateInt64(),
			"liquid staked shares for validator %s", moniker)
	}
}
