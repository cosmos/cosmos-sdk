package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abonormal scenarios are not tested here.
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
		InitialStake: 1000,
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var stakingGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &stakingGenesis)

	require.Equal(t, uint32(207), stakingGenesis.Params.MaxValidators)
	require.Equal(t, uint32(7), stakingGenesis.Params.MaxEntries)
	require.Equal(t, uint32(48), stakingGenesis.Params.HistoricalEntries)
	require.Equal(t, "stake", stakingGenesis.Params.BondDenom)
	require.Equal(t, float64(238280), stakingGenesis.Params.UnbondingTime.Seconds())
	// check numbers of Delegations and Validators
	require.Len(t, stakingGenesis.Delegations, 3)
	require.Len(t, stakingGenesis.Validators, 3)
	// check Delegations
	require.Equal(t, "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", stakingGenesis.Delegations[0].DelegatorAddress)
	require.Equal(t, "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", stakingGenesis.Delegations[0].ValidatorAddress)
	require.Equal(t, "1000.000000000000000000", stakingGenesis.Delegations[0].Shares.String())
	// check validators
	require.Equal(t, "cosmosvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnsvnaes", stakingGenesis.Validators[2].GetOperator().String())
	require.Equal(t, "cosmosvalconspub1zcjduepq280tm686ma80cva9z620dmknd9a858pd2zmq9ackfenfllecjxds0hg9n7", stakingGenesis.Validators[2].ConsensusPubkey)
	require.Equal(t, false, stakingGenesis.Validators[2].Jailed)
	require.Equal(t, "BOND_STATUS_UNBONDED", stakingGenesis.Validators[2].Status.String())
	require.Equal(t, "1000", stakingGenesis.Validators[2].Tokens.String())
	require.Equal(t, "1000.000000000000000000", stakingGenesis.Validators[2].DelegatorShares.String())
	require.Equal(t, "0.292059246265731326", stakingGenesis.Validators[2].Commission.CommissionRates.Rate.String())
	require.Equal(t, "0.330000000000000000", stakingGenesis.Validators[2].Commission.CommissionRates.MaxRate.String())
	require.Equal(t, "0.038337453731274481", stakingGenesis.Validators[2].Commission.CommissionRates.MaxChangeRate.String())
	require.Equal(t, "1", stakingGenesis.Validators[2].MinSelfDelegation.String())
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
			}, "invalid memory address or nil pointer dereference"},
		{
			// panic => reason: numBonded != len(Accnounts)
			module.SimulationState{
				AppParams:    make(simtypes.AppParams),
				Cdc:          cdc,
				Rand:         r,
				NumBonded:    4,
				Accounts:     simtypes.RandomAccounts(r, 3),
				InitialStake: 1000,
				GenState:     make(map[string]json.RawMessage),
			}, "invalid memory address or nil pointer dereference"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}
