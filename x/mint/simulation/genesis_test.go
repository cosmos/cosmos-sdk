package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/mint/simulation"
	"cosmossdk.io/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abonormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          encCfg.Codec,
		Rand:         r,
		NumBonded:    3,
		BondDenom:    sdk.DefaultBondDenom,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: math.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var mintGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &mintGenesis)

	require.Equal(t, "0.170000000000000000", mintGenesis.Minter.Inflation.String())
	require.Equal(t, "0.000000000000000000", mintGenesis.Minter.AnnualProvisions.String())
	require.Equal(t, "1970-01-01 00:00:00 +0000 UTC", mintGenesis.Minter.GenesisTime.String())
	require.Nil(t, mintGenesis.Minter.PreviousBlockTime)
	require.Equal(t, "stake", mintGenesis.Minter.MintDenom)
}

// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})

	s := rand.NewSource(1)
	r := rand.New(s)
	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       encCfg.Codec,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		tt := tt

		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}
