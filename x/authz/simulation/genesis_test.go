package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz/testutil"
)

func TestRandomizedGenState(t *testing.T) {
	var cdc codec.Codec
	depinject.Inject(testutil.AppConfig, &cdc)

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
	authzGenesis, ok := simState.GenState[authz.ModuleName].(*authz.GenesisState)
	require.True(t, ok)

	require.Len(t, authzGenesis.Authorization, len(simState.Accounts)-1)
}
