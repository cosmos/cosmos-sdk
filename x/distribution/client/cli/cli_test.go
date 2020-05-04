// +build cli_test

package cli_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint"
)

func TestCLIWithdrawRewards(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	genesisState := f.GenesisState()
	inflationMin := sdk.MustNewDecFromStr("1.0")
	var mintData mint.GenesisState
	f.Cdc.UnmarshalJSON(genesisState[mint.ModuleName], &mintData)
	mintData.Minter.Inflation = inflationMin
	mintData.Params.InflationMin = inflationMin
	mintData.Params.InflationMax = sdk.MustNewDecFromStr("1.0")
	mintDataBz, err := f.Cdc.MarshalJSON(mintData)
	require.NoError(t, err)
	genesisState[mint.ModuleName] = mintDataBz

	genFile := filepath.Join(f.SimdHome, "config", "genesis.json")
	genDoc, err := tmtypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	genDoc.AppState, err = f.Cdc.MarshalJSON(genesisState)
	require.NoError(t, genDoc.SaveAs(genFile))

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	fooAddr := f.KeyAddress(cli.KeyFoo)
	rewards := testutil.QueryRewards(f, fooAddr)
	require.Equal(t, 1, len(rewards.Rewards))
	require.NotNil(t, rewards.Total)

	fooVal := sdk.ValAddress(fooAddr)
	success := testutil.TxWithdrawRewards(f, fooVal, fooAddr.String(), "-y")
	require.True(t, success)

	rewards = testutil.QueryRewards(f, fooAddr)
	require.Equal(t, 1, len(rewards.Rewards))

	require.Nil(t, rewards.Total)
	f.Cleanup()
}
