package simapp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// genutil delivers gentxs through the router during InitChain, so x/safety
// must init after genutil. A chain whose genesis disables every msg must still
// create its validators from gentxs and produce blocks.
func TestSafetyGenesisDisabledAllWithGentxs(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1

	var gs safetytypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[safetytypes.ModuleName], &gs)
	gs.DisabledPatterns = []string{"*"}
	cfg.GenesisState[safetytypes.ModuleName] = cfg.Codec.MustMarshalJSON(&gs)

	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	defer net.Cleanup()

	_, err = net.WaitForHeight(2)
	require.NoError(t, err)
}
