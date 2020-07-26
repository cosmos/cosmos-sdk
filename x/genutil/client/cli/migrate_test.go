package cli

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/testutil"
)

func TestGetMigrationCallback(t *testing.T) {
	for _, version := range GetMigrationVersions() {
		require.NotNil(t, GetMigrationCallback(version))
	}
}

func TestMigrateGenesis(t *testing.T) {
	home, cleanup := testutil.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	cdc := makeCodec()

	genesisPath := path.Join(home, "genesis.json")
	target := "v0.36"

	cmd := MigrateGenesisCmd()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)

	clientCtx := client.Context{}.WithJSONMarshaler(cdc)
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
