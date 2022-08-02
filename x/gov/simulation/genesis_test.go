package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abnormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]proto.Message),
	}

	simulation.RandomizedGenState(&simState)

	govGenesis, ok := simState.GenState[types.ModuleName].(*v1.GenesisState)
	require.True(t, ok)

	dec1, _ := sdk.NewDecFromStr("0.361000000000000000")
	dec2, _ := sdk.NewDecFromStr("0.512000000000000000")
	dec3, _ := sdk.NewDecFromStr("0.267000000000000000")

	require.Equal(t, "905stake", govGenesis.Params.MinDeposit[0].String())
	require.Equal(t, "77h26m10s", govGenesis.Params.MaxDepositPeriod.String())
	require.Equal(t, float64(148296), govGenesis.Params.VotingPeriod.Seconds())
	require.Equal(t, dec1.String(), govGenesis.Params.Quorum)
	require.Equal(t, dec2.String(), govGenesis.Params.Threshold)
	require.Equal(t, dec3.String(), govGenesis.Params.VetoThreshold)
	require.Equal(t, uint64(0x28), govGenesis.StartingProposalId)
	require.Equal(t, []*v1.Deposit(nil), govGenesis.Deposits)
	require.Equal(t, []*v1.Vote(nil), govGenesis.Votes)
	require.Equal(t, []*v1.Proposal(nil), govGenesis.Proposals)
}

// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

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
