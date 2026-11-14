package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abnormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState, simulation.RandomGenesisAccounts)

	var authGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &authGenesis)

	require.Equal(t, uint64(0x8c), authGenesis.Params.GetMaxMemoCharacters())
	require.Equal(t, uint64(0x2b6), authGenesis.Params.GetSigVerifyCostED25519())
	require.Equal(t, uint64(0x1ff), authGenesis.Params.GetSigVerifyCostSecp256k1())
	require.Equal(t, uint64(9), authGenesis.Params.GetTxSigLimit())
	require.Equal(t, uint64(5), authGenesis.Params.GetTxSizeCostPerByte())
	require.Equal(t, false, authGenesis.Params.GetPubKeyChangeEnabled())
	require.Equal(t, uint64(0x4928), authGenesis.Params.GetPubKeyChangeCost())

	genAccounts, err := types.UnpackAccounts(authGenesis.Accounts)
	require.NoError(t, err)
	require.Len(t, genAccounts, 3)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", genAccounts[2].GetAddress().String())
	require.Equal(t, uint64(0), genAccounts[2].GetAccountNumber())
	require.Equal(t, uint64(0), genAccounts[2].GetSequence())
}

// TestRandomizedGenStatePubKeyChange checks that the account rekeying params
// are randomized: both values of PubKeyChangeEnabled occur across seeds, and
// every generated PubKeyChangeCost is valid.
func TestRandomizedGenStatePubKeyChange(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	seen := map[bool]bool{}
	for seed := int64(0); seed < 20; seed++ {
		r := rand.New(rand.NewSource(seed))
		simState := module.SimulationState{
			AppParams:    make(simtypes.AppParams),
			Cdc:          cdc,
			Rand:         r,
			NumBonded:    3,
			Accounts:     simtypes.RandomAccounts(r, 3),
			InitialStake: sdkmath.NewInt(1000),
			GenState:     make(map[string]json.RawMessage),
		}
		simulation.RandomizedGenState(&simState, simulation.RandomGenesisAccounts)

		var authGenesis types.GenesisState
		simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &authGenesis)
		require.NoError(t, authGenesis.Params.Validate())
		require.NotEqual(t, types.DefaultPubKeyChangeCost, authGenesis.Params.PubKeyChangeCost)
		seen[authGenesis.Params.PubKeyChangeEnabled] = true
	}
	require.True(t, seen[true], "pub_key_change_enabled never randomized to true")
	require.True(t, seen[false], "pub_key_change_enabled never randomized to false")
}
