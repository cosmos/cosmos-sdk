package cli_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestGetMigrationCallback(t *testing.T) {
	for _, version := range cli.GetMigrationVersions() {
		require.NotNil(t, cli.GetMigrationCallback(version))
	}
}

func TestMigrateGenesis(t *testing.T) {
	home := t.TempDir()

	cdc := makeCodec()

	genesisPath := path.Join(home, "genesis.json")
	target := "v0.36"

	cmd := cli.MigrateGenesisCmd()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)

	clientCtx := client.Context{}.WithLegacyAmino(cdc)
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	// Reject if we dont' have the right parameters or genesis does not exists
	cmd.SetArgs([]string{target, genesisPath})
	require.Error(t, cmd.ExecuteContext(ctx))

	// Noop migration with minimal genesis
	emptyGenesis := []byte(`{"chain_id":"test","app_state":{}}`)
	require.NoError(t, ioutil.WriteFile(genesisPath, emptyGenesis, 0644))

	cmd.SetArgs([]string{target, genesisPath})
	require.NoError(t, cmd.ExecuteContext(ctx))
}

func TestSanitizeTendermintGenesis(t *testing.T) {
	// An example exported genesis file from a 0.37 chain. Note that evidence
	// parameters only contains `max_age`.
	v037Exported := []byte(`{
  "app_hash": "",
  "app_state": {},
  "chain_id": "test",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1",
      "time_iota_ms": "1000"
    },
    "evidence": { "max_age": "100000" },
    "validator": { "pub_key_types": ["ed25519"] }
  },
  "genesis_time": "2020-09-29T20:16:29.172362037Z",
  "validators": []
}`)

	// We expect an error decoding an older `consensus_params` with the latest
	// TM.
	_, err := tmtypes.GenesisDocFromJSON(v037Exported)
	require.Error(t, err)

	_, err := cli.SanitizeTendermintGenesis(v037Exported)
	require.NoError(t, err)

	require.True(t, false)
}
