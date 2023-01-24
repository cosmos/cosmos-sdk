package simulation_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/simulation"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

func TestRandomizedGenState(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          simapp.MakeTestEncodingConfig().Codec,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 10),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	var err error
	bankGenBefore := banktypes.GenesisState{}
	simState.GenState[banktypes.ModuleName], err = simState.Cdc.MarshalJSON(&bankGenBefore)
	require.NoError(t, err, "MarshalJSON empty bank genesis state")

	fundsHolder := authtypes.NewModuleAddress(quarantine.ModuleName)

	simulation.RandomizedGenState(&simState, fundsHolder)
	var gen quarantine.GenesisState
	err = simState.Cdc.UnmarshalJSON(simState.GenState[quarantine.ModuleName], &gen)
	require.NoError(t, err, "UnmarshalJSON on quarantine genesis state")

	totalQuarantined := sdk.Coins{}
	for _, qf := range gen.QuarantinedFunds {
		totalQuarantined = totalQuarantined.Add(qf.Coins...)
	}

	if !totalQuarantined.IsZero() {
		var bankGen banktypes.GenesisState
		err = simState.Cdc.UnmarshalJSON(simState.GenState[banktypes.ModuleName], &bankGen)
		require.NoError(t, err, "UnmarshalJSON on quarantine bank state")
		holder := fundsHolder.String()
		holderFound := false
		for _, bal := range bankGen.Balances {
			if holder == bal.Address {
				holderFound = true
				assert.Equal(t, totalQuarantined.String(), bal.Coins.String())
			}
		}
		assert.True(t, holderFound, "no balance entry found for the funds holder")
		_, hasNeg := bankGen.Supply.SafeSub(totalQuarantined...)
		assert.False(t, hasNeg, "not enough supply %s to cover the total quarantined %s", bankGen.Supply.String(), totalQuarantined.String())
	}
}

func TestRandomizedGenStateImportExport(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 10)
	emptyBankGen := banktypes.GenesisState{}
	emptyBankGenBz, err := cdc.MarshalJSON(&emptyBankGen)
	require.NoError(t, err, "MarshalJSON empty bank genesis state")
	fundsHolder := authtypes.NewModuleAddress(quarantine.ModuleName)

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
			simState.GenState[banktypes.ModuleName] = emptyBankGenBz

			simulation.RandomizedGenState(&simState, fundsHolder)
			var randomGenState quarantine.GenesisState
			err = simState.Cdc.UnmarshalJSON(simState.GenState[quarantine.ModuleName], &randomGenState)
			require.NoError(t, err, "UnmarshalJSON on quarantine genesis state")

			// The unspecified auto-responses don't get written, so we need to remove them to get the expected.
			expGenState := MakeCopyOfGenesisState(&randomGenState)
			expectedAutoResponses := make([]*quarantine.AutoResponseEntry, 0, len(expGenState.AutoResponses))
			for _, entry := range expGenState.AutoResponses {
				if entry.Response != quarantine.AUTO_RESPONSE_UNSPECIFIED {
					expectedAutoResponses = append(expectedAutoResponses, entry)
				}
			}
			expGenState.AutoResponses = expectedAutoResponses

			var bankGen banktypes.GenesisState
			err = simState.Cdc.UnmarshalJSON(simState.GenState[banktypes.ModuleName], &bankGen)
			require.NoError(t, err, "UnmarshalJSON on bank genesis state")

			app := simapp.Setup(t, false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{})

			testBankInit := func() {
				app.BankKeeper.InitGenesis(ctx, &bankGen)
			}
			require.NotPanics(t, testBankInit, "bank InitGenesis")

			testInit := func() {
				app.QuarantineKeeper.InitGenesis(ctx, &randomGenState)
			}
			require.NotPanics(t, testInit, "quarantine InitGenesis")

			var actualGenState *quarantine.GenesisState
			testExport := func() {
				actualGenState = app.QuarantineKeeper.ExportGenesis(ctx)
			}
			require.NotPanics(t, testExport, "ExportGenesis")

			// Note: The contents of the genesis state is not expected to be in the same order after the init/export.
			// I could probably go through the trouble of sorting things, but it would either be horribly inefficient or annoyingly complex (probably both).
			// Primarily, the genesis state uses bech32 encoding for the addresses, but when exported, the entries are sorted based on their byte values.
			// And sorting by bech32 does not equal sorting by byte values.
			assert.ElementsMatch(t, expGenState.QuarantinedAddresses, actualGenState.QuarantinedAddresses, "QuarantinedAddresses, A = expected, B = actual")
			assert.ElementsMatch(t, expGenState.AutoResponses, actualGenState.AutoResponses, "AutoResponses, A = expected, B = actual")
			assert.ElementsMatch(t, expGenState.QuarantinedFunds, actualGenState.QuarantinedFunds, "QuarantinedFunds, A = expected, B = actual")
		})
		if !passed {
			break
		}
	}
}

func TestRandomQuarantinedAddresses(t *testing.T) {
	// Once RandomAccounts is called, we can't trust the values returned from r.
	// So all we can do here is check the length of the returned list using seed values found through trial and error.
	// These will probably be prone to breakage since any change in use of r will alter the outcomes.
	// In the event that this test fails, make sure that there was a change that should alter the outcomes.
	// If you've verified that use of r has changed, you can look at the logs of the " good seeds" test to get the
	// new expected seed values for each entry.

	type testCase struct {
		name   string
		seed   int64
		expLen int
	}

	tests := []*testCase{
		{
			name:   "zero",
			seed:   103,
			expLen: 0,
		},
		{
			name:   "one",
			seed:   3,
			expLen: 1,
		},
		{
			name:   "two",
			seed:   17,
			expLen: 2,
		},
		{
			name:   "three",
			seed:   2,
			expLen: 3,
		},
		{
			name:   "four",
			seed:   4,
			expLen: 4,
		},
		{
			name:   "five",
			seed:   0,
			expLen: 5,
		},
		{
			name:   "six",
			seed:   15,
			expLen: 6,
		},
		{
			name:   "seven",
			seed:   31,
			expLen: 7,
		},
		{
			name:   "eight",
			seed:   45,
			expLen: 8,
		},
		{
			name:   "nine",
			seed:   238,
			expLen: 9,
		},
	}

	runTest := func(t *testing.T, tc *testCase) bool {
		t.Helper()
		rv := true
		// Using a separate rand to generate accounts to make it easier to predict the func being tested.
		accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(1)), tc.expLen*4)
		r := rand.New(rand.NewSource(tc.seed))
		actual := simulation.RandomQuarantinedAddresses(r, accounts)
		if assert.Len(t, actual, tc.expLen, "QuarantinedAddresses") {
			if tc.expLen == 0 {
				rv = assert.Nil(t, actual, "QuarantinedAddresses") && rv
			}
		} else {
			rv = false
		}
		for i, addr := range actual {
			assert.NotEmpty(t, addr, "QuarantinedAddress[%d]", i)
		}
		return rv
	}

	allPassed := true
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allPassed = runTest(t, tc) && allPassed
		})
	}

	if !allPassed {
		stillBad := make(map[string]bool)
		maxAttempts := 10000
		t.Run("find good seeds", func(t *testing.T) {
			for _, tc := range tests {
				for i := 0; i < maxAttempts; i++ {
					if runTest(t, tc) {
						break
					}
					tc.seed += 1
				}
			}
		})
		// opening space is on purpose so it gets sorted to the top.
		t.Run(" good seeds", func(t *testing.T) {
			for _, tc := range tests {
				if stillBad[tc.name] {
					t.Logf("%q => no passing seed found from %d to %d", tc.name, int(tc.seed)-maxAttempts, tc.seed-1)
				} else {
					t.Logf("%q => %d", tc.name, tc.seed)
				}
			}
			t.Fail() // Only runs if the whole test fails. Marking this subtest as failed draws attention to it.
		})
	}
}

func TestRandomQuarantineAutoResponses(t *testing.T) {
	// Once RandomAccounts is called, we can't trust the values returned from r.
	// In here, using seeds found through trial and error, we can check that some
	// addrs are kept, others ignored, and some new ones added.
	// These will probably be prone to breakage since any change in use of r will alter the outcomes.
	// In the event that this test fails, make sure that there was a change that should alter the outcomes.
	// If you've verified that use of r has changed, you can look at the logs of the " good seeds" test to get the
	// new expected seed values for each entry.

	type testCase struct {
		name     string
		seed     int64
		qAddrs   []string
		expAddrs []string
		newAddrs int
	}

	tests := []*testCase{
		{
			name:     "no addrs in no new addrs",
			seed:     3,
			qAddrs:   nil,
			expAddrs: nil,
			newAddrs: 0,
		},
		{
			name:     "no addrs in one new addr",
			seed:     1,
			qAddrs:   nil,
			expAddrs: nil,
			newAddrs: 1,
		},
		{
			name:     "one addr in is kept",
			seed:     5,
			qAddrs:   []string{"addr1"},
			expAddrs: []string{"addr1"},
			newAddrs: 0,
		},
		{
			name:     "one addr in is not kept",
			seed:     4,
			qAddrs:   []string{"addr1"},
			expAddrs: nil,
			newAddrs: 0,
		},
		{
			name:     "two addrs in first kept new added",
			seed:     2,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{"addr1"},
			newAddrs: 1,
		},
	}

	runTest := func(t *testing.T, tc *testCase) bool {
		t.Helper()
		rv := true
		r := rand.New(rand.NewSource(tc.seed))
		actual := simulation.RandomQuarantineAutoResponses(r, tc.qAddrs)
		addrMap := make(map[string]bool)
		for i, entry := range actual {
			addrMap[entry.ToAddress] = true
			assert.NotEmpty(t, entry.ToAddress, "[%d].ToAddress", i)
			assert.NotEmpty(t, entry.FromAddress, "[%d].FromAddress", i)
			assert.True(t, entry.Response.IsValid(), "[%d].Response.IsValid(), Response = %s", i, entry.Response)
		}
		addrs := make([]string, 0, len(addrMap))
		for addr := range addrMap {
			addrs = append(addrs, addr)
		}
		rv = assert.Len(t, addrs, len(tc.expAddrs)+tc.newAddrs, "to addresses") && rv
		for _, addr := range tc.expAddrs {
			rv = assert.Contains(t, addrs, addr, "to addresses") && rv
		}
		return rv
	}

	allPassed := true
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allPassed = runTest(t, tc) && allPassed
		})
	}

	if !allPassed {
		stillBad := make(map[string]bool)
		maxAttempts := 10000
		t.Run("find good seeds", func(t *testing.T) {
			for _, tc := range tests {
				for i := 0; i < maxAttempts; i++ {
					if runTest(t, tc) {
						break
					}
					tc.seed += 1
				}
			}
		})
		// opening space is on purpose so it gets sorted to the top.
		t.Run(" good seeds", func(t *testing.T) {
			for _, tc := range tests {
				if stillBad[tc.name] {
					t.Logf("%q => no passing seed found from %d to %d", tc.name, int(tc.seed)-maxAttempts, tc.seed-1)
				} else {
					t.Logf("%q => %d", tc.name, tc.seed)
				}
			}
			t.Fail() // Only runs if the whole test fails. Marking this subtest as failed draws attention to it.
		})
	}
}

func TestRandomQuarantinedFunds(t *testing.T) {
	// Once RandomAccounts is called, we can't trust the values returned from r.
	// In here, using seeds found through trial and error, we can check that some
	// addrs are kept, others ignored.
	// These will probably be prone to breakage since any change in use of r will alter the outcomes.
	// In the event that this test fails, make sure that there was a change that should alter the outcomes.
	// If you've verified that use of r has changed, you can look at the logs of the " good seeds" test to get the
	// new expected seed values for each entry.

	type testCase struct {
		name     string
		seed     int64
		qAddrs   []string
		expAddrs []string
	}

	tests := []*testCase{
		{
			name:     "no addrs in",
			seed:     0,
			qAddrs:   nil,
			expAddrs: nil,
		},
		{
			name:     "one addr in is kept",
			seed:     0,
			qAddrs:   []string{"addr1"},
			expAddrs: []string{"addr1"},
		},
		{
			name:     "one addr in is not kept",
			seed:     3,
			qAddrs:   []string{"addr1"},
			expAddrs: nil,
		},
		{
			name:     "two addrs in none kept",
			seed:     8,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{},
		},
		{
			name:     "two addrs in first kept",
			seed:     4,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{"addr1"},
		},
		{
			name:     "two addrs in second kept",
			seed:     3,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{"addr2"},
		},
		{
			name:     "two addrs in both kept",
			seed:     0,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{"addr1", "addr2"},
		},
	}

	runTest := func(t *testing.T, tc *testCase) bool {
		t.Helper()
		rv := true
		r := rand.New(rand.NewSource(tc.seed))
		actual := simulation.RandomQuarantinedFunds(r, tc.qAddrs)
		addrMap := make(map[string]bool)
		for i, entry := range actual {
			addrMap[entry.ToAddress] = true
			assert.NotEmpty(t, entry.ToAddress, "[%d].ToAddress", i)
			for j, addr := range entry.UnacceptedFromAddresses {
				assert.NotEmpty(t, addr, "[%d].UnacceptedFromAddresses[%d]", i, j)
			}
			assert.NoError(t, entry.Coins.Validate(), "[%d].Coins", i)
		}
		addrs := make([]string, 0, len(addrMap))
		for addr := range addrMap {
			addrs = append(addrs, addr)
		}
		rv = assert.Len(t, addrs, len(tc.expAddrs), "to addresses") && rv
		for _, addr := range tc.expAddrs {
			rv = assert.Contains(t, addrs, addr, "to addresses") && rv
		}
		return rv
	}

	allPassed := true
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allPassed = runTest(t, tc) && allPassed
		})
	}

	if !allPassed {
		stillBad := make(map[string]bool)
		maxAttempts := 10000
		t.Run("find good seeds", func(t *testing.T) {
			for _, tc := range tests {
				for i := 0; i < maxAttempts; i++ {
					if runTest(t, tc) {
						break
					}
					tc.seed += 1
				}
			}
		})
		// opening space is on purpose so it gets sorted to the top.
		t.Run(" good seeds", func(t *testing.T) {
			for _, tc := range tests {
				if stillBad[tc.name] {
					t.Logf("%q => no passing seed found from %d to %d", tc.name, int(tc.seed)-maxAttempts, tc.seed-1)
				} else {
					t.Logf("%q => %d", tc.name, tc.seed)
				}
			}
			t.Fail() // Only runs if the whole test fails. Marking this subtest as failed draws attention to it.
		})
	}
}
