package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/cosmos/cosmos-sdk/x/nft/simulation"
	"github.com/cosmos/cosmos-sdk/x/nft/testutil"
)

func TestRandomizedGenState(t *testing.T) {
	cfg, err := network.DefaultConfigWithAppConfig(testutil.AppConfig)
	require.NoError(t, err)

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cfg.Codec,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)
	var nftGenesis nft.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[nft.ModuleName], &nftGenesis)

	require.Len(t, nftGenesis.Classes, len(simState.Accounts)-1)
	require.Len(t, nftGenesis.Entries, len(simState.Accounts)-1)
}
