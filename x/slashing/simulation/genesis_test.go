package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abonormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	cdc := codec.New()
	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: 1000,
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var stakingGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &stakingGenesis)

	dec1, _ := sdk.NewDecFromStr("0.600000000000000000")
	dec2, _ := sdk.NewDecFromStr("0.022222222222222222")
	dec3, _ := sdk.NewDecFromStr("0.008928571428571429")

	require.Equal(t, dec1, stakingGenesis.Params.MinSignedPerWindow)
	require.Equal(t, dec2, stakingGenesis.Params.SlashFractionDoubleSign)
	require.Equal(t, dec3, stakingGenesis.Params.SlashFractionDowntime)
	require.Equal(t, int64(720), stakingGenesis.Params.SignedBlocksWindow)
	require.Equal(t, time.Duration(34800000000000), stakingGenesis.Params.DowntimeJailDuration)
	require.Len(t, stakingGenesis.MissedBlocks, 0)
	require.Len(t, stakingGenesis.SigningInfos, 0)

}

// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
	cdc := codec.New()

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
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}
