package cli_test

import (
	"github.com/cosmos/cosmos-sdk/tests/cli"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrcli "github.com/cosmos/cosmos-sdk/x/distribution/client/cli_test"
	"github.com/cosmos/cosmos-sdk/x/mint"
)

func TestCliWithdrawRewards(t *testing.T) {
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
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(cli.KeyFoo)
	rewards := distrcli.QueryRewards(f, fooAddr)
	require.Equal(t, 1, len(rewards.Rewards))
	require.NotNil(t, rewards.Total)

	fooVal := sdk.ValAddress(fooAddr)
	success := distrcli.TxWithdrawRewards(f, fooVal, fooAddr.String(), "-y")
	require.True(t, success)

	rewards = distrcli.QueryRewards(f, fooAddr)
	require.Equal(t, 1, len(rewards.Rewards))

	require.Nil(t, rewards.Total)
	f.Cleanup()
}
