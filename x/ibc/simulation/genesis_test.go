package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/simulation"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abonormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	cdc := codec.New()
	s := rand.NewSource(1)
	r := rand.New(s)

	// Make sure to register cdc.
	// Otherwise RandomizedGenState will panic!
	types.RegisterCodec(cdc)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: 1000,
		GenState:     make(map[string]json.RawMessage),
	}

	// Remark: the current RandomizedGenState function
	// is actually not random as it does not utilize concretely the random value r.
	// This tests will pass for any value of r.
	simulation.RandomizedGenState(&simState)

	var ibcGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[host.ModuleName], &ibcGenesis)

	require.Len(t, ibcGenesis.Clients, 0)
	require.Len(t, ibcGenesis.ClientsConsensus, 0)
	require.False(t, ibcGenesis.CreateLocalhost)

	// Note: ibcGenesis.ChannelGenesis is missing because the ChannelGenesis
	// interface is not register in RegisterCodec. (Not sure if is a feature
	// or a bug.)

}

// TestRandomizedGenState tests the execution of RandomizedGenState
// without registering the IBC client interfaces and types.
// We expect the test to panic.
func TestRandomizedGenState1(t *testing.T) {
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

	require.Panicsf(t, func() { simulation.RandomizedGenState(&simState) }, "failed to marshal JSON: Unregistered interface exported.ClientState")

}
