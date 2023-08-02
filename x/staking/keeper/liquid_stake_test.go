package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// Helper function to create a base account from an account name
// Used to differentiate against liquid staking provider module account
func createBaseAccount(app *simapp.SimApp, ctx sdk.Context, accountName string) sdk.AccAddress {
	baseAccountAddress := sdk.AccAddress(accountName)
	app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccountWithAddress(baseAccountAddress))
	return baseAccountAddress
}

// Helper function to create 32-length account
// Used to mock an liquid staking provider's ICA account
func createICAAccount(app *simapp.SimApp, ctx sdk.Context) sdk.AccAddress {
	icahost := "icahost"
	connectionID := "connection-0"
	portID := icahost

	moduleAddress := authtypes.NewModuleAddress(icahost)
	icaAddress := sdk.AccAddress(address.Derive(moduleAddress, []byte(connectionID+portID)))

	account := authtypes.NewBaseAccountWithAddress(icaAddress)
	app.AccountKeeper.SetAccount(ctx, account)

	return icaAddress
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

// Tests Set/Get TotalLiquidStakedTokens
func TestTotalLiquidStakedTokens(t *testing.T) {
	_, app, ctx := createTestInput()

	// Update the total liquid staked
	total := sdk.NewInt(100)
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, total)

	// Confirm it was updated
	require.Equal(t, total, app.StakingKeeper.GetTotalLiquidStakedTokens(ctx), "initial")
}

// Tests Increase/Decrease TotalValidatorTotalLiquidShares
func TestValidatorTotalLiquidShares(t *testing.T) {
	_, app, ctx := createTestInput()

	// Create a validator address
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	// Set an initial total
	initial := sdk.NewDec(100)
	validator := types.Validator{
		OperatorAddress:   valAddress.String(),
		TotalLiquidShares: initial,
	}
	app.StakingKeeper.SetValidator(ctx, validator)
}

// Tests DelegatorIsLiquidStaker
func TestDelegatorIsLiquidStaker(t *testing.T) {
	_, app, ctx := createTestInput()

	// Create base and ICA accounts
	baseAccountAddress := createBaseAccount(app, ctx, "base-account")
	icaAccountAddress := createICAAccount(app, ctx)

	// Only the ICA module account should be considered a liquid staking provider
	require.False(t, app.StakingKeeper.DelegatorIsLiquidStaker(baseAccountAddress), "base account")
	require.True(t, app.StakingKeeper.DelegatorIsLiquidStaker(icaAccountAddress), "ICA module account")
}

// Helper function to clear the Bonded pool balances before a unit test
func clearPoolBalance(t *testing.T, app *simapp.SimApp, ctx sdk.Context) {
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	initialBondedBalance := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(types.BondedPoolName), bondDenom)

	err := app.BankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, minttypes.ModuleName, sdk.NewCoins(initialBondedBalance))
	require.NoError(t, err, "no error expected when clearing bonded pool balance")
}

// Helper function to fund the Bonded pool balances before a unit test
func fundPoolBalance(t *testing.T, app *simapp.SimApp, ctx sdk.Context, amount sdk.Int) {
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	bondedPoolCoin := sdk.NewCoin(bondDenom, amount)

	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(bondedPoolCoin))
	require.NoError(t, err, "no error expected when minting")

	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.BondedPoolName, sdk.NewCoins(bondedPoolCoin))
	require.NoError(t, err, "no error expected when sending tokens to bonded pool")
}

// Tests CheckExceedsGlobalLiquidStakingCap
func TestCheckExceedsGlobalLiquidStakingCap(t *testing.T) {
	_, app, ctx := createTestInput()

	testCases := []struct {
		name             string
		globalLiquidCap  sdk.Dec
		totalLiquidStake sdk.Int
		totalStake       sdk.Int
		newLiquidStake   sdk.Int
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
			params := app.StakingKeeper.GetParams(ctx)
			params.GlobalLiquidStakingCap = tc.globalLiquidCap
			app.StakingKeeper.SetParams(ctx, params)

			// Update the total liquid tokens
			app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, tc.totalLiquidStake)

			// Fund each pool for the given test case
			clearPoolBalance(t, app, ctx)
			fundPoolBalance(t, app, ctx, tc.totalStake)

			// Check if the new tokens would exceed the global cap
			actualExceeds := app.StakingKeeper.CheckExceedsGlobalLiquidStakingCap(ctx, tc.newLiquidStake, tc.tokenizingShares)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests SafelyIncreaseTotalLiquidStakedTokens
func TestSafelyIncreaseTotalLiquidStakedTokens(t *testing.T) {
	_, app, ctx := createTestInput()

	intitialTotalLiquidStaked := sdk.NewInt(100)
	increaseAmount := sdk.NewInt(10)
	poolBalance := sdk.NewInt(200)

	// Set the total staked and total liquid staked amounts
	// which are required components when checking the global cap
	// Total stake is calculated from the pool balance
	clearPoolBalance(t, app, ctx)
	fundPoolBalance(t, app, ctx, poolBalance)
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)

	// Set the global cap such that a small delegation would exceed the cap
	params := app.StakingKeeper.GetParams(ctx)
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.0001")
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the total liquid stake again, it should error since
	// the cap was exceeded
	err := app.StakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount, true)
	require.ErrorIs(t, err, types.ErrGlobalLiquidStakingCapExceeded)
	require.Equal(t, intitialTotalLiquidStaked, app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))

	// Now relax the cap so that the increase succeeds
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.99")
	app.StakingKeeper.SetParams(ctx, params)

	// Confirm the total increased
	err = app.StakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount, true)
	require.NoError(t, err)
	require.Equal(t, intitialTotalLiquidStaked.Add(increaseAmount), app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))
}

// Tests DecreaseTotalLiquidStakedTokens
func TestDecreaseTotalLiquidStakedTokens(t *testing.T) {
	_, app, ctx := createTestInput()

	intitialTotalLiquidStaked := sdk.NewInt(100)
	decreaseAmount := sdk.NewInt(10)

	// Set the total liquid staked to an arbitrary value
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)

	// Decrease the total liquid stake and confirm the total was updated
	err := app.StakingKeeper.DecreaseTotalLiquidStakedTokens(ctx, decreaseAmount)
	require.NoError(t, err, "no error expected when decreasing total liquid staked tokens")
	require.Equal(t, intitialTotalLiquidStaked.Sub(decreaseAmount), app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))

	// Attempt to decrease by an excessive amount, it should error
	err = app.StakingKeeper.DecreaseTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)
	require.ErrorIs(t, err, types.ErrTotalLiquidStakedUnderflow)
}

// Tests CheckExceedsValidatorBondCap
func TestCheckExceedsValidatorBondCap(t *testing.T) {
	_, app, ctx := createTestInput()

	testCases := []struct {
		name                string
		validatorShares     sdk.Dec
		validatorBondFactor sdk.Dec
		currentLiquidShares sdk.Dec
		newShares           sdk.Dec
		expectedExceeds     bool
	}{
		{
			// Validator Shares: 100, Factor: 1, Current Shares: 90 => 100 Max Shares, Capacity: 10
			// New Shares: 5 - below cap
			name:                "factor 1 - below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(1),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(5),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 1, Current Shares: 90 => 100 Max Shares, Capacity: 10
			// New Shares: 10 - at cap
			name:                "factor 1 - at cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(1),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(10),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 1, Current Shares: 90 => 100 Max Shares, Capacity: 10
			// New Shares: 15 - above cap
			name:                "factor 1 - above cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(1),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(15),
			expectedExceeds:     true,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 5 - below cap
			name:                "factor 2 - well below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(5),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 100 - below cap
			name:                "factor 2 - below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(100),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 110 - below cap
			name:                "factor 2 - at cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(110),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 111 - above cap
			name:                "factor 2 - above cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(111),
			expectedExceeds:     true,
		},
		{
			// Validator Shares: 100, Factor: 100, Current Shares: 90 => 10000 Max Shares, Capacity: 9910
			// New Shares: 100 - below cap
			name:                "factor 100 - below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(100),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(100),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 100, Current Shares: 90 => 10000 Max Shares, Capacity: 9910
			// New Shares: 9910 - at cap
			name:                "factor 100 - at cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(100),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(9910),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 100, Current Shares: 90 => 10000 Max Shares, Capacity: 9910
			// New Shares: 9911 - above cap
			name:                "factor 100 - above cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(100),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(9911),
			expectedExceeds:     true,
		},
		{
			// Factor of -1 (disabled): Should always return false
			name:                "factor disabled",
			validatorShares:     sdk.NewDec(1),
			validatorBondFactor: sdk.NewDec(-1),
			currentLiquidShares: sdk.NewDec(1),
			newShares:           sdk.NewDec(1_000_000),
			expectedExceeds:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the validator bond factor
			params := app.StakingKeeper.GetParams(ctx)
			params.ValidatorBondFactor = tc.validatorBondFactor
			app.StakingKeeper.SetParams(ctx, params)

			// Create a validator with designated self-bond shares
			validator := types.Validator{
				TotalLiquidShares:        tc.currentLiquidShares,
				TotalValidatorBondShares: tc.validatorShares,
			}

			// Check whether the cap is exceeded
			actualExceeds := app.StakingKeeper.CheckExceedsValidatorBondCap(ctx, validator, tc.newShares)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests TestCheckExceedsValidatorLiquidStakingCap
func TestCheckExceedsValidatorLiquidStakingCap(t *testing.T) {
	_, app, ctx := createTestInput()

	testCases := []struct {
		name                  string
		validatorLiquidCap    sdk.Dec
		validatorLiquidShares sdk.Dec
		validatorTotalShares  sdk.Dec
		newLiquidShares       sdk.Dec
		expectedExceeds       bool
	}{
		{
			// Cap: 10% - Delegation Below Threshold
			// Liquid Shares: 5, Total Shares: 95, New Liquid Shares: 1
			// => Liquid Shares: 5+1=6, Total Shares: 95+1=96 => 6/96 = 6% < 10% cap
			name:                  "10 percent cap _ delegation below cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.1"),
			validatorLiquidShares: sdk.NewDec(5),
			validatorTotalShares:  sdk.NewDec(95),
			newLiquidShares:       sdk.NewDec(1),
			expectedExceeds:       false,
		},
		{
			// Cap: 10% - Delegation At Threshold
			// Liquid Shares: 5, Total Shares: 95, New Liquid Shares: 5
			// => Liquid Shares: 5+5=10, Total Shares: 95+5=100 => 10/100 = 10% == 10% cap
			name:                  "10 percent cap _ delegation equals cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.1"),
			validatorLiquidShares: sdk.NewDec(5),
			validatorTotalShares:  sdk.NewDec(95),
			newLiquidShares:       sdk.NewDec(4),
			expectedExceeds:       false,
		},
		{
			// Cap: 10% - Delegation Exceeds Threshold
			// Liquid Shares: 5, Total Shares: 95, New Liquid Shares: 6
			// => Liquid Shares: 5+6=11, Total Shares: 95+6=101 => 11/101 = 11% > 10% cap
			name:                  "10 percent cap _ delegation exceeds cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.1"),
			validatorLiquidShares: sdk.NewDec(5),
			validatorTotalShares:  sdk.NewDec(95),
			newLiquidShares:       sdk.NewDec(6),
			expectedExceeds:       true,
		},
		{
			// Cap: 20% - Delegation Below Threshold
			// Liquid Shares: 20, Total Shares: 220, New Liquid Shares: 29
			// => Liquid Shares: 20+29=49, Total Shares: 220+29=249 => 49/249 = 19% < 20% cap
			name:                  "20 percent cap _ delegation below cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.2"),
			validatorLiquidShares: sdk.NewDec(20),
			validatorTotalShares:  sdk.NewDec(220),
			newLiquidShares:       sdk.NewDec(29),
			expectedExceeds:       false,
		},
		{
			// Cap: 20% - Delegation At Threshold
			// Liquid Shares: 20, Total Shares: 220, New Liquid Shares: 30
			// => Liquid Shares: 20+30=50, Total Shares: 220+30=250 => 50/250 = 20% == 20% cap
			name:                  "20 percent cap _ delegation equals cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.2"),
			validatorLiquidShares: sdk.NewDec(20),
			validatorTotalShares:  sdk.NewDec(220),
			newLiquidShares:       sdk.NewDec(30),
			expectedExceeds:       false,
		},
		{
			// Cap: 20% - Delegation Exceeds Threshold
			// Liquid Shares: 20, Total Shares: 220, New Liquid Shares: 31
			// => Liquid Shares: 20+31=51, Total Shares: 220+31=251 => 51/251 = 21% > 20% cap
			name:                  "20 percent cap _ delegation exceeds cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.2"),
			validatorLiquidShares: sdk.NewDec(20),
			validatorTotalShares:  sdk.NewDec(220),
			newLiquidShares:       sdk.NewDec(31),
			expectedExceeds:       true,
		},
		{
			// Cap of 0% - everything should exceed
			name:                  "0 percent cap",
			validatorLiquidCap:    sdk.ZeroDec(),
			validatorLiquidShares: sdk.NewDec(0),
			validatorTotalShares:  sdk.NewDec(1_000_000),
			newLiquidShares:       sdk.NewDec(1),
			expectedExceeds:       true,
		},
		{
			// Cap of 100% - nothing should exceed
			name:                  "100 percent cap",
			validatorLiquidCap:    sdk.OneDec(),
			validatorLiquidShares: sdk.NewDec(1),
			validatorTotalShares:  sdk.NewDec(1_000_000),
			newLiquidShares:       sdk.NewDec(1),
			expectedExceeds:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the validator liquid staking cap
			params := app.StakingKeeper.GetParams(ctx)
			params.ValidatorLiquidStakingCap = tc.validatorLiquidCap
			app.StakingKeeper.SetParams(ctx, params)

			// Create a validator with designated self-bond shares
			validator := types.Validator{
				TotalLiquidShares: tc.validatorLiquidShares,
				DelegatorShares:   tc.validatorTotalShares,
			}

			// Check whether the cap is exceeded
			actualExceeds := app.StakingKeeper.CheckExceedsValidatorLiquidStakingCap(ctx, validator, tc.newLiquidShares)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests SafelyIncreaseValidatorTotalLiquidShares
func TestSafelyIncreaseValidatorTotalLiquidShares(t *testing.T) {
	_, app, ctx := createTestInput()

	// Generate a test validator address
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	// Helper function to check the validator's liquid shares
	checkValidatorLiquidShares := func(expected sdk.Dec, description string) {
		actualValidator, found := app.StakingKeeper.GetValidator(ctx, valAddress)
		require.True(t, found)
		require.Equal(t, expected.TruncateInt64(), actualValidator.TotalLiquidShares.TruncateInt64(), description)
	}

	// Start with the following:
	//   Initial Liquid Shares: 0
	//   Validator Bond Shares: 10
	//   Validator TotalShares: 75
	//
	// Initial Caps:
	//   ValidatorBondFactor: 1 (Cap applied at 10 shares)
	//   ValidatorLiquidStakingCap: 25% (Cap applied at 25 shares)
	//
	// Cap Increases:
	//   ValidatorBondFactor: 10 (Cap applied at 100 shares)
	//   ValidatorLiquidStakingCap: 40% (Cap applied at 50 shares)
	initialLiquidShares := sdk.NewDec(0)
	validatorBondShares := sdk.NewDec(10)
	validatorTotalShares := sdk.NewDec(75)

	firstIncreaseAmount := sdk.NewDec(20)
	secondIncreaseAmount := sdk.NewDec(40)

	initialBondFactor := sdk.NewDec(1)
	finalBondFactor := sdk.NewDec(10)
	initialLiquidStakingCap := sdk.MustNewDecFromStr("0.25")
	finalLiquidStakingCap := sdk.MustNewDecFromStr("0.4")

	// Create a validator with designated self-bond shares
	initialValidator := types.Validator{
		OperatorAddress:          valAddress.String(),
		TotalLiquidShares:        initialLiquidShares,
		TotalValidatorBondShares: validatorBondShares,
		DelegatorShares:          validatorTotalShares,
	}
	app.StakingKeeper.SetValidator(ctx, initialValidator)

	// Set validator bond factor to a small number such that any delegation would fail,
	// and set the liquid staking cap such that the first stake would succeed, but the second
	// would fail
	params := app.StakingKeeper.GetParams(ctx)
	params.ValidatorBondFactor = initialBondFactor
	params.ValidatorLiquidStakingCap = initialLiquidStakingCap
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the validator liquid shares, it should throw an
	// error that the validator bond cap was exceeded
	validator := initialValidator
	err := app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, &validator, firstIncreaseAmount)
	require.ErrorIs(t, err, types.ErrInsufficientValidatorBondShares)
	checkValidatorLiquidShares(initialLiquidShares, "shares after low bond factor")

	// Change validator bond factor to a more conservative number, so that the increase succeeds
	params.ValidatorBondFactor = finalBondFactor
	app.StakingKeeper.SetParams(ctx, params)

	// Try the increase again and check that it succeeded
	validator = initialValidator
	expectedLiquidSharesAfterFirstStake := initialLiquidShares.Add(firstIncreaseAmount)
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, &validator, firstIncreaseAmount)
	require.NoError(t, err)
	checkValidatorLiquidShares(expectedLiquidSharesAfterFirstStake, "shares with cap loose bond cap")

	// Attempt another increase, it should fail from the liquid staking cap
	validator = initialValidator
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, &validator, secondIncreaseAmount)
	require.ErrorIs(t, err, types.ErrValidatorLiquidStakingCapExceeded)
	checkValidatorLiquidShares(expectedLiquidSharesAfterFirstStake, "shares after liquid staking cap hit")

	// Raise the liquid staking cap so the new increment succeeds
	params.ValidatorLiquidStakingCap = finalLiquidStakingCap
	app.StakingKeeper.SetParams(ctx, params)

	// Finally confirm that the increase succeeded this time
	validator = initialValidator
	expectedLiquidSharesAfterSecondStake := initialLiquidShares.Add(secondIncreaseAmount)
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, &validator, secondIncreaseAmount)
	require.NoError(t, err, "no error expected after increasing liquid staking cap")
	checkValidatorLiquidShares(expectedLiquidSharesAfterSecondStake, "shares after loose liquid stake cap")
}

// Tests DecreaseValidatorTotalLiquidShares
func TestDecreaseValidatorTotalLiquidShares(t *testing.T) {
	_, app, ctx := createTestInput()

	initialLiquidShares := sdk.NewDec(100)
	decreaseAmount := sdk.NewDec(10)

	// Create a validator with designated self-bond shares
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	initialValidator := types.Validator{
		OperatorAddress:   valAddress.String(),
		TotalLiquidShares: initialLiquidShares,
	}
	app.StakingKeeper.SetValidator(ctx, initialValidator)

	// Decrease the validator liquid shares, and confirm the new share amount has been updated
	err := app.StakingKeeper.DecreaseValidatorTotalLiquidShares(ctx, &initialValidator, decreaseAmount)
	require.NoError(t, err, "no error expected when decreasing validator total liquid shares")

	actualValidator, found := app.StakingKeeper.GetValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, initialLiquidShares.Sub(decreaseAmount), actualValidator.TotalLiquidShares, "total liquid shares")

	// Attempt to decrease by a larger amount than it has, it should fail
	err = app.StakingKeeper.DecreaseValidatorTotalLiquidShares(ctx, &actualValidator, initialLiquidShares)
	require.ErrorIs(t, err, types.ErrValidatorLiquidSharesUnderflow)
}

// Tests SafelyDecreaseValidatorBond
func TestSafelyDecreaseValidatorBond(t *testing.T) {
	_, app, ctx := createTestInput()

	// Initial Bond Factor: 100, Initial Validator Bond: 10
	// => Max Liquid Shares 1000 (Initial Liquid Shares: 200)
	initialBondFactor := sdk.NewDec(100)
	initialValidatorBondShares := sdk.NewDec(10)
	initialLiquidShares := sdk.NewDec(200)

	// Create a validator with designated self-bond shares
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	initialValidator := types.Validator{
		OperatorAddress:          valAddress.String(),
		TotalValidatorBondShares: initialValidatorBondShares,
		TotalLiquidShares:        initialLiquidShares,
	}
	app.StakingKeeper.SetValidator(ctx, initialValidator)

	// Set the bond factor
	params := app.StakingKeeper.GetParams(ctx)
	params.ValidatorBondFactor = initialBondFactor
	app.StakingKeeper.SetParams(ctx, params)

	// Decrease the validator bond from 10 to 5 (minus 5)
	// This will adjust the cap (factor * shares)
	// from (100 * 10 = 1000) to (100 * 5 = 500)
	// Since this is still above the initial liquid shares of 200, this will succeed
	decreaseAmount, expectedBondShares := sdk.NewDec(5), sdk.NewDec(5)
	err := app.StakingKeeper.SafelyDecreaseValidatorBond(ctx, &initialValidator, decreaseAmount)
	require.NoError(t, err)

	actualValidator, found := app.StakingKeeper.GetValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, expectedBondShares, actualValidator.TotalValidatorBondShares, "validator bond shares shares")

	// Now attempt to decrease the validator bond again from 5 to 1 (minus 4)
	// This time, the cap will be reduced to (factor * shares) = (100 * 1) = 100
	// However, the total liquid shares are currently 200, so this should fail
	decreaseAmount, expectedBondShares = sdk.NewDec(4), sdk.NewDec(1)
	validator := actualValidator
	err = app.StakingKeeper.SafelyDecreaseValidatorBond(ctx, &validator, decreaseAmount)
	require.ErrorIs(t, err, types.ErrInsufficientValidatorBondShares)

	// Finally, disable the cap and attempt to decrease again
	// This time it should succeed
	params.ValidatorBondFactor = types.ValidatorBondCapDisabled
	app.StakingKeeper.SetParams(ctx, params)

	validator = actualValidator
	err = app.StakingKeeper.SafelyDecreaseValidatorBond(ctx, &validator, decreaseAmount)
	require.NoError(t, err)

	actualValidator, found = app.StakingKeeper.GetValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, expectedBondShares, actualValidator.TotalValidatorBondShares, "validator bond shares shares")
}

// Tests Add/Remove/Get/SetTokenizeSharesLock
func TestTokenizeSharesLock(t *testing.T) {
	_, app, ctx := createTestInput()

	addresses := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1))
	addressA, addressB := addresses[0], addresses[1]

	unlocked := types.TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED.String()
	locked := types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED.String()
	lockExpiring := types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING.String()

	// Confirm both accounts start unlocked
	status, _ := app.StakingKeeper.GetTokenizeSharesLock(ctx, addressA)
	require.Equal(t, unlocked, status.String(), "addressA unlocked at start")

	status, _ = app.StakingKeeper.GetTokenizeSharesLock(ctx, addressB)
	require.Equal(t, unlocked, status.String(), "addressB unlocked at start")

	// Lock the first account
	app.StakingKeeper.AddTokenizeSharesLock(ctx, addressA)

	// The first account should now have tokenize shares disabled
	// and the unlock time should be the zero time
	status, _ = app.StakingKeeper.GetTokenizeSharesLock(ctx, addressA)
	require.Equal(t, locked, status.String(), "addressA locked")

	status, _ = app.StakingKeeper.GetTokenizeSharesLock(ctx, addressB)
	require.Equal(t, unlocked, status.String(), "addressB still unlocked")

	// Update the lock time and confirm it was set
	expectedUnlockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	app.StakingKeeper.SetTokenizeSharesUnlockTime(ctx, addressA, expectedUnlockTime)

	status, actualUnlockTime := app.StakingKeeper.GetTokenizeSharesLock(ctx, addressA)
	require.Equal(t, lockExpiring, status.String(), "addressA lock expiring")
	require.Equal(t, expectedUnlockTime, actualUnlockTime, "addressA unlock time")

	// Confirm B is still unlocked
	status, _ = app.StakingKeeper.GetTokenizeSharesLock(ctx, addressB)
	require.Equal(t, unlocked, status.String(), "addressB still unlocked")

	// Remove the lock
	app.StakingKeeper.RemoveTokenizeSharesLock(ctx, addressA)
	status, _ = app.StakingKeeper.GetTokenizeSharesLock(ctx, addressA)
	require.Equal(t, unlocked, status.String(), "addressA unlocked at end")

	status, _ = app.StakingKeeper.GetTokenizeSharesLock(ctx, addressB)
	require.Equal(t, unlocked, status.String(), "addressB unlocked at end")
}

// Tests GetAllTokenizeSharesLocks
func TestGetAllTokenizeSharesLocks(t *testing.T) {
	_, app, ctx := createTestInput()

	addresses := simapp.AddTestAddrs(app, ctx, 4, sdk.NewInt(1))

	// Set 2 locked accounts, and two accounts with a lock expiring
	app.StakingKeeper.AddTokenizeSharesLock(ctx, addresses[0])
	app.StakingKeeper.AddTokenizeSharesLock(ctx, addresses[1])

	unlockTime1 := time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC)
	unlockTime2 := time.Date(2023, 1, 2, 1, 0, 0, 0, time.UTC)
	app.StakingKeeper.SetTokenizeSharesUnlockTime(ctx, addresses[2], unlockTime1)
	app.StakingKeeper.SetTokenizeSharesUnlockTime(ctx, addresses[3], unlockTime2)

	// Defined expected locks after GetAll
	expectedLocks := map[string]types.TokenizeShareLock{
		addresses[0].String(): {
			Status: types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED.String(),
		},
		addresses[1].String(): {
			Status: types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED.String(),
		},
		addresses[2].String(): {
			Status:         types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING.String(),
			CompletionTime: unlockTime1,
		},
		addresses[3].String(): {
			Status:         types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING.String(),
			CompletionTime: unlockTime2,
		},
	}

	// Check output from GetAll
	actualLocks := app.StakingKeeper.GetAllTokenizeSharesLocks(ctx)
	require.Len(t, actualLocks, len(expectedLocks), "number of locks")

	for i, actual := range actualLocks {
		expected, ok := expectedLocks[actual.Address]
		require.True(t, ok, "address %s not expected", actual.Address)
		require.Equal(t, expected.Status, actual.Status, "tokenize share lock #%d status", i)
		require.Equal(t, expected.CompletionTime, actual.CompletionTime, "tokenize share lock #%d completion time", i)
	}
}

// Test Get/SetPendingTokenizeShareAuthorizations
func TestPendingTokenizeShareAuthorizations(t *testing.T) {
	_, app, ctx := createTestInput()

	// Create dummy accounts and completion times
	addresses := simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1))
	addressStrings := []string{}
	for _, address := range addresses {
		addressStrings = append(addressStrings, address.String())
	}

	timeA := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	timeB := timeA.Add(time.Hour)

	// There should be no addresses returned originally
	authorizationsA := app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeA)
	require.Empty(t, authorizationsA.Addresses, "no addresses at timeA expected")

	authorizationsB := app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeB)
	require.Empty(t, authorizationsB.Addresses, "no addresses at timeB expected")

	// Store addresses for timeB
	app.StakingKeeper.SetPendingTokenizeShareAuthorizations(ctx, timeB, types.PendingTokenizeShareAuthorizations{
		Addresses: addressStrings,
	})

	// Check addresses
	authorizationsA = app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeA)
	require.Empty(t, authorizationsA.Addresses, "no addresses at timeA expected at end")

	authorizationsB = app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeB)
	require.Equal(t, addressStrings, authorizationsB.Addresses, "address length")
}

// Test QueueTokenizeSharesAuthorization and RemoveExpiredTokenizeShareLocks
func TestTokenizeShareAuthorizationQueue(t *testing.T) {
	_, app, ctx := createTestInput()

	// We'll start by adding the following addresses to the queue
	//   Time 0: [address0]
	//   Time 1: []
	//   Time 2: [address1, address2, address3]
	//   Time 3: [address4, address5]
	//   Time 4: [address6]
	addresses := simapp.AddTestAddrs(app, ctx, 7, sdk.NewInt(1))
	addressesByTime := map[int][]sdk.AccAddress{
		0: {addresses[0]},
		1: {},
		2: {addresses[1], addresses[2], addresses[3]},
		3: {addresses[4], addresses[5]},
		4: {addresses[6]},
	}

	// Set the unbonding time to 1 day
	unbondingPeriod := time.Hour * 24
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = unbondingPeriod
	app.StakingKeeper.SetParams(ctx, params)

	// Add each address to the queue and then increment the block time
	// such that the times line up as follows
	//   Time 0: 2023-01-01 00:00:00
	//   Time 1: 2023-01-01 00:01:00
	//   Time 2: 2023-01-01 00:02:00
	//   Time 3: 2023-01-01 00:03:00
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(startTime)
	blockTimeIncrement := time.Hour

	for timeIndex := 0; timeIndex <= 4; timeIndex++ {
		for _, address := range addressesByTime[timeIndex] {
			app.StakingKeeper.QueueTokenizeSharesAuthorization(ctx, address)
		}
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(blockTimeIncrement))
	}

	// We'll unlock the tokens using the following progression
	// The "alias'"/keys for these times assume a starting point of the Time 0
	// from above, plus the Unbonding Time
	//   Time -1  (2023-01-01 23:59:99): []
	//   Time  0  (2023-01-02 00:00:00): [address0]
	//   Time  1  (2023-01-02 00:01:00): []
	//   Time 2.5 (2023-01-02 00:02:30): [address1, address2, address3]
	//   Time 10  (2023-01-02 00:10:00): [address4, address5, address6]
	unlockBlockTimes := map[string]time.Time{
		"-1":  startTime.Add(unbondingPeriod).Add(-time.Second),
		"0":   startTime.Add(unbondingPeriod),
		"1":   startTime.Add(unbondingPeriod).Add(blockTimeIncrement),
		"2.5": startTime.Add(unbondingPeriod).Add(2 * blockTimeIncrement).Add(blockTimeIncrement / 2),
		"10":  startTime.Add(unbondingPeriod).Add(10 * blockTimeIncrement),
	}
	expectedUnlockedAddresses := map[string][]string{
		"-1":  {},
		"0":   {addresses[0].String()},
		"1":   {},
		"2.5": {addresses[1].String(), addresses[2].String(), addresses[3].String()},
		"10":  {addresses[4].String(), addresses[5].String(), addresses[6].String()},
	}

	// Now we'll remove items from the queue sequentially
	// First check with a block time before the first expiration - it should remove no addresses
	actualAddresses := app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["-1"])
	require.Equal(t, expectedUnlockedAddresses["-1"], actualAddresses, "no addresses unlocked from time -1")

	// Then pass in (time 0 + unbonding time) - it should remove the first address
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["0"])
	require.Equal(t, expectedUnlockedAddresses["0"], actualAddresses, "one address unlocked from time 0")

	// Now pass in (time 1 + unbonding time) - it should remove no addresses since
	// the address at time 0 was already removed
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["1"])
	require.Equal(t, expectedUnlockedAddresses["1"], actualAddresses, "no addresses unlocked from time 1")

	// Now pass in (time 2.5 + unbonding time) - it should remove the three addresses from time 2
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["2.5"])
	require.Equal(t, expectedUnlockedAddresses["2.5"], actualAddresses, "addresses unlocked from time 2.5")

	// Finally pass in a block time far in the future, which should remove all the remaining locks
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["10"])
	require.Equal(t, expectedUnlockedAddresses["10"], actualAddresses, "addresses unlocked from time 10")
}

// Test RefreshTotalLiquidStaked
func TestRefreshTotalLiquidStaked(t *testing.T) {
	_, app, ctx := createTestInput()

	// Set an arbitrary total liquid staked tokens amount that will get overwritten by the refresh
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, sdk.NewInt(999))

	// Add validator's with various exchange rates
	validators := []types.Validator{
		{
			// Exchange rate of 1
			OperatorAddress:   "valA",
			Tokens:            sdk.NewInt(100),
			DelegatorShares:   sdk.NewDec(100),
			TotalLiquidShares: sdk.NewDec(100), // should be overwritten
		},
		{
			// Exchange rate of 0.9
			OperatorAddress:   "valB",
			Tokens:            sdk.NewInt(90),
			DelegatorShares:   sdk.NewDec(100),
			TotalLiquidShares: sdk.NewDec(200), // should be overwritten
		},
		{
			// Exchange rate of 0.75
			OperatorAddress:   "valC",
			Tokens:            sdk.NewInt(75),
			DelegatorShares:   sdk.NewDec(100),
			TotalLiquidShares: sdk.NewDec(300), // should be overwritten
		},
	}

	// Add various delegations across the above validator's
	// Total Liquid Staked: 1,849 + 922 = 2,771
	// Total Liquid Shares:
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
			delegatorAddress = createICAAccount(app, ctx)
		case delegationCase.isTokenized:
			delegatorAddress = createTokenizeShareModuleAccount(1)
		default:
			delegatorAddress = createBaseAccount(app, ctx, delegationCase.delegation.DelegatorAddress)
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

		actualLiquidShares := actualValidator.TotalLiquidShares
		require.Equal(t, expectedLiquidShares.TruncateInt64(), actualLiquidShares.TruncateInt64(),
			"liquid staked shares for validator %s", moniker)
	}
}
