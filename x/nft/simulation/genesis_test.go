package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/nft"                  //nolint:staticcheck // deprecated and to be removed
	nftmodule "github.com/cosmos/cosmos-sdk/x/nft/module" //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/nft/simulation"
)

func TestRandomizedGenState(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(nftmodule.AppModuleBasic{})

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          encCfg.Codec,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState, addresscodec.NewBech32Codec("cosmos"))
	var nftGenesis nft.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[nft.ModuleName], &nftGenesis)

	require.Len(t, nftGenesis.Classes, len(simState.Accounts)-1)
	require.Len(t, nftGenesis.Entries, len(simState.Accounts)-1)
}
