package simulation_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/simulation"
)

// getAccountCount gets the number of accounts to use.
func getAccountCount() int {
	// This is a function instead of variable because I dislike global variables.
	// Global variables in unit tests just leads to weird, painful problems.
	return 10
}

// newDefaultRand returns a new Rand made from a default, constant seed value.
func newDefaultRand() *rand.Rand {
	return rand.New(rand.NewSource(1))
}

// generateAccounts generates a standard number of accounts.
func generateAccounts(t *testing.T) []simtypes.Account {
	// This uses its own random number generator in order to not affect the numbers
	// generated elsewhere. Plus, by setting the seed the same, we should always
	// have the same accounts which can help when trying to identify problems.
	rv := simtypes.RandomAccounts(newDefaultRand(), getAccountCount())
	// Log all the account addresses so that if this test breaks,
	// people can see which accounts where which.
	t.Logf("accounts:")
	for i, acct := range rv {
		t.Logf("[%d]: %s", i, acct.Address.String())
	}
	return rv
}

func TestRandomizer(t *testing.T) {
	// These tests serve primarily to show you what random numbers are being picked,
	// but also to alert you if that ever changes.
	// If it changes, several other tests in this file will fail too.
	// If others fail, but this test does not, it's not because the
	// generator is making different numbers, but it could still be due to
	// a change in use of the generator (i.e. r is being used differently).

	// The expectedRands values in here were defined by running the test once and seeing what came out.

	// randomSanctionedAddressesRands generates the random numbers used in RandomSanctionedAddresses
	randomSanctionedAddressesRands := func(r *rand.Rand) []int64 {
		// RandomSanctionedAddresses uses r.Int63n(5) for each account.
		rv := make([]int64, getAccountCount())
		for i := range rv {
			rv[i] = r.Int63n(5)
		}
		return rv
	}

	t.Run("values used in RandomSanctionedAddresses", func(t *testing.T) {
		expectedRands := []int64{0, 1, 1, 1, 2, 0, 3, 3, 1, 4}

		r := newDefaultRand()
		actualRands := randomSanctionedAddressesRands(r)
		assert.Equal(t, expectedRands, actualRands, "random numbers generated")
	})

	// randomTempEntriesRands generates the random numbers used in RandomTempEntries
	randomTempEntriesRands := func(r *rand.Rand) []int64 {
		// RandomTempEntries uses r.Int63n(10) to decide which account to use.
		// If 0 or 1 then r.Int63n(1000) is called for the prop id.
		rv := make([]int64, 0, getAccountCount()*2)
		for i := 0; i < 10; i++ {
			v := r.Int63n(10)
			rv = append(rv, v)
			if v <= 1 {
				rv = append(rv, r.Int63n(1000))
			}
		}
		return rv
	}

	t.Run("values used in RandomTempEntries", func(t *testing.T) {
		expectedRands := []int64{0, 551, 1, 51, 7, 0, 758, 8, 6, 9, 4, 7, 4}

		r := newDefaultRand()
		actualRands := randomTempEntriesRands(r)
		assert.Equal(t, expectedRands, actualRands, "random numbers generated")
	})

	// randomParamsRands generates the random numbers used in RandomParams
	randomParamsRands := func(r *rand.Rand) []int64 {
		// RandomParams uses r.Int63n(5) twice.
		// For each, if not 0, then r.Int63n(1_000) is used.
		rv := make([]int64, 0, 4)
		for i := 0; i < 2; i++ {
			v := r.Int63n(5)
			rv = append(rv, v)
			if v != 0 {
				rv = append(rv, r.Int63n(1000))
			}
		}
		return rv
	}

	t.Run("values used in RandomParams default seed", func(t *testing.T) {
		expectedRands := []int64{0, 1, 821}

		r := newDefaultRand()
		actualRands := randomParamsRands(r)
		assert.Equal(t, expectedRands, actualRands, "random numbers generated")
	})

	t.Run("values used in RandomParams seed 100", func(t *testing.T) {
		// A little crossover here. This knowledge is useful in the operations tests for updating params.
		expectedRands := []int64{3, 24, 4, 39}

		r := rand.New(rand.NewSource(100))
		actualRands := randomParamsRands(r)
		assert.Equal(t, expectedRands, actualRands, "random numbers generated")
	})

	t.Run("values used in immediate operations for various seeds", func(t *testing.T) {
		// A little crossover here. This knowledge is useful in the operations tests for immediate stuff.
		expectedRands := []int{0, 1}
		actualRands := make([]int, len(expectedRands))
		for i := range actualRands {
			actualRands[i] = rand.New(rand.NewSource(int64(i))).Intn(2)
		}
		assert.Equal(t, expectedRands, actualRands, "first random number generated at each seed")
	})
}

func TestRandomSanctionedAddresses(t *testing.T) {
	accounts := generateAccounts(t)

	// From TestRandomizer: 0, 1, 1, 1, 2, 0, 3, 3, 1, 4
	expected := []string{
		accounts[0].Address.String(),
		accounts[5].Address.String(),
	}

	r := newDefaultRand()
	var actual []string
	testFunc := func() {
		actual = simulation.RandomSanctionedAddresses(r, accounts)
	}
	require.NotPanics(t, testFunc, "RandomSanctionedAddresses")
	assert.Equal(t, expected, actual, "RandomSanctionedAddresses result")
}

func TestRandomTempEntries(t *testing.T) {
	accounts := generateAccounts(t)

	// From TestRandomizer: 0, 551, 1, 51, 7, 0, 758, 8, 6, 9, 4, 7, 4
	expected := []*sanction.TemporaryEntry{
		{
			Address:    accounts[0].Address.String(),
			ProposalId: 551 + 1_000_000_000,
			Status:     sanction.TEMP_STATUS_SANCTIONED,
		},
		{
			Address:    accounts[1].Address.String(),
			ProposalId: 51 + 2_000_000_000,
			Status:     sanction.TEMP_STATUS_UNSANCTIONED,
		},
		{
			Address:    accounts[3].Address.String(),
			ProposalId: 758 + 1_000_000_000,
			Status:     sanction.TEMP_STATUS_SANCTIONED,
		},
	}

	r := newDefaultRand()
	var actual []*sanction.TemporaryEntry
	testFunc := func() {
		actual = simulation.RandomTempEntries(r, accounts)
	}
	require.NotPanics(t, testFunc, "RandomTempEntries")
	assert.Equal(t, expected, actual, "RandomTempEntries result")
}

func TestRandomParams(t *testing.T) {
	// From TestRandomizer: 0, 1, 821
	expected := &sanction.Params{
		ImmediateSanctionMinDeposit:   nil,
		ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 821+1)),
	}

	r := newDefaultRand()
	var actual *sanction.Params
	testFunc := func() {
		actual = simulation.RandomParams(r)
	}
	require.NotPanics(t, testFunc, "RandomParams")
	assert.Equal(t, expected, actual, "RandomParams result")
}

func TestRandomizedGenState(t *testing.T) {
	accounts := generateAccounts(t)
	accountMap := make(map[string]bool)
	for _, acc := range accounts {
		accountMap[acc.Address.String()] = true
	}

	// Since part of RandomizedGenState involves creating new accounts, and that uses r before anything else can use it,
	// it's not worthwhile to define a specific expected state.
	// Instead, I just need to make sure that the addresses included aren't in the accounts provided.

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          simapp.MakeTestEncodingConfig().Codec,
		Rand:         rand.New(rand.NewSource(0)), // not using getSeedValue because that was used to generate the accounts.
		NumBonded:    3,
		Accounts:     accounts,
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	testFunc := func() {
		simulation.RandomizedGenState(&simState)
	}
	require.NotPanics(t, testFunc, "RandomizedGenState")

	sanctionGenStateBz := simState.GenState[sanction.ModuleName]
	var actual sanction.GenesisState
	err := simState.Cdc.UnmarshalJSON(sanctionGenStateBz, &actual)
	require.NoError(t, err, "UnmarshalJSON to sanction.GenesisState")

	for i, addr := range actual.SanctionedAddresses {
		isKnownAcc := accountMap[addr]
		assert.False(t, isKnownAcc, "is SanctionedAddresses[%d] a known account", i)
	}
	for i, entry := range actual.TemporaryEntries {
		isKnownAcc := accountMap[entry.Address]
		assert.False(t, isKnownAcc, "is TemporaryEntries[%d] a known account", i)
	}
	// This is kind of useless because UnmarshalJSON will create an empty Params no matter what, but....
	assert.NotNil(t, actual.Params, "Params")
	// Not asserting any param values though since their randomization depends on uses of r outside our control.
	// They'd be prone to breakage for seemingly unrelated reasons, which is dumb and annoying.

	// Make sure nothing else was added to the genesis state.
	expectedGenStateKeys := []string{sanction.ModuleName}
	var actualGenStateKeys []string
	for key := range simState.GenState {
		actualGenStateKeys = append(actualGenStateKeys, key)
	}
	sort.Strings(actualGenStateKeys)
	assert.Equal(t, expectedGenStateKeys, actualGenStateKeys, "keys in GenState")
}

func TestRandomizedGenStateImportExport(t *testing.T) {
	// This goes through 1001 seeds and:
	// 1. generates a random genesis,
	// 2. imports it into an app,
	// 3. exports the sanction genesis state from the app,
	// 4. makes sure the exported gen state is equal to the one randomly generated.
	// It will stop at the first seed that fails.

	cdc := simapp.MakeTestEncodingConfig().Codec
	accounts := generateAccounts(t)

	for i := int64(0); i <= 1000; i++ {
		passed := t.Run(fmt.Sprintf("seed %d", i), func(t *testing.T) {
			simState := module.SimulationState{
				AppParams:    make(simtypes.AppParams),
				Cdc:          cdc,
				Rand:         rand.New(rand.NewSource(i)),
				NumBonded:    3,
				Accounts:     make([]simtypes.Account, len(accounts)),
				InitialStake: sdkmath.NewInt(1000),
				GenState:     make(map[string]json.RawMessage),
			}
			copy(simState.Accounts, accounts)

			// Generate the random genesis state.
			testRandGen := func() {
				simulation.RandomizedGenState(&simState)
			}
			require.NotPanics(t, testRandGen, "RandomizedGenState")

			// Extract the randomly generated genesis state from the simState.
			var randomGenState sanction.GenesisState
			err := simState.Cdc.UnmarshalJSON(simState.GenState[sanction.ModuleName], &randomGenState)
			require.NoError(t, err, "UnmarshalJSON to sanction.GenesisState")

			// Set the Coins in Params to nil if they're zero/empty.
			// That's how they'll come back from the export.
			if randomGenState.Params.ImmediateSanctionMinDeposit.IsZero() {
				randomGenState.Params.ImmediateSanctionMinDeposit = nil
			}
			if randomGenState.Params.ImmediateUnsanctionMinDeposit.IsZero() {
				randomGenState.Params.ImmediateUnsanctionMinDeposit = nil
			}

			// Create a new app, so we can init + export.
			app := simapp.Setup(t, false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{})

			// Do the init on the random genesis state.
			testInit := func() {
				app.SanctionKeeper.InitGenesis(ctx, &randomGenState)
			}
			require.NotPanics(t, testInit, "sanction InitGenesis")

			// Export the app's sanction genesis state.
			var actualGenState *sanction.GenesisState
			testExport := func() {
				actualGenState = app.SanctionKeeper.ExportGenesis(ctx)
			}
			require.NotPanics(t, testExport, "ExportGenesis")

			// Note: The contents of the genesis states is not expected to be in the same order after the init/export.
			// I could probably go through the trouble of sorting things, but it would either be horribly inefficient or annoyingly complex (probably both).
			// Primarily, the genesis state uses bech32 encoding for the addresses, but when exported, the entries are sorted based on their byte values.
			// And sorting by bech32 does not equal sorting by byte values.
			assert.ElementsMatch(t, randomGenState.SanctionedAddresses, actualGenState.SanctionedAddresses, "SanctionedAddresses, A = expected, B = actual")
			assert.ElementsMatch(t, randomGenState.TemporaryEntries, actualGenState.TemporaryEntries, "TemporaryEntries, A = expected, B = actual")
			if !assert.Equal(t, randomGenState.Params, actualGenState.Params, "Params") && randomGenState.Params != nil && actualGenState.Params != nil {
				assert.Equal(t, randomGenState.Params.ImmediateSanctionMinDeposit.String(),
					actualGenState.Params.ImmediateSanctionMinDeposit.String(),
					"Params.ImmediateSanctionMinDeposit")
				assert.Equal(t, randomGenState.Params.ImmediateUnsanctionMinDeposit.String(),
					actualGenState.Params.ImmediateUnsanctionMinDeposit.String(),
					"Params.ImmediateUnsanctionMinDeposit")
			}
		})
		if !passed {
			break
		}
	}
}
