package cli_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

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
	home, cleanup := testutil.NewTestCaseDir(t)
	t.Cleanup(cleanup)

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
