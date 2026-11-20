package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/safety/simulation"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

// RandomizedGenState writes a valid genesis that keeps the given authority,
// draws the operator from the simulation accounts or leaves it empty, and
// starts with an empty disabled list.
func TestRandomizedGenState(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	ac := addresscodec.NewBech32Codec("cosmos")
	gov := sdk.AccAddress(address.Module("gov")).String()
	r := rand.New(rand.NewSource(1))
	accounts := simtypes.RandomAccounts(r, 3)
	addrs := make([]string, len(accounts))
	for i, acc := range accounts {
		addrs[i] = acc.Address.String()
	}

	var sawOperator, sawNoOperator, sawScope, sawNoScope bool
	for range 50 {
		simState := module.SimulationState{
			Cdc:      cdc,
			Rand:     r,
			GenState: make(map[string]json.RawMessage),
			Accounts: accounts,
		}
		simulation.RandomizedGenState(&simState, gov)

		var gs types.GenesisState
		cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &gs)
		require.NoError(t, gs.Validate(ac))
		require.Equal(t, gov, gs.Params.Authority)
		require.Empty(t, gs.DisabledPatterns)
		if gs.Params.Operator != "" {
			require.Contains(t, addrs, gs.Params.Operator)
		}
		sawOperator = sawOperator || gs.Params.Operator != ""
		sawNoOperator = sawNoOperator || gs.Params.Operator == ""
		sawScope = sawScope || len(gs.Params.Scope) > 0
		sawNoScope = sawNoScope || len(gs.Params.Scope) == 0
	}
	require.True(t, sawOperator && sawNoOperator && sawScope && sawNoScope)

	// with no accounts the operator stays empty
	simState := module.SimulationState{Cdc: cdc, Rand: r, GenState: make(map[string]json.RawMessage)}
	simulation.RandomizedGenState(&simState, gov)
	var gs types.GenesisState
	cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &gs)
	require.NoError(t, gs.Validate(ac))
	require.Empty(t, gs.Params.Operator)
}
