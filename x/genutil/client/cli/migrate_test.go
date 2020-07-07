package cli

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/tendermint/tendermint/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
)

func setupCmd(genesisTime string, chainID string) *cobra.Command {
	c := &cobra.Command{
		Use:  "c",
		Args: cobra.ArbitraryArgs,
		Run:  func(_ *cobra.Command, args []string) {},
	}

	c.Flags().String(flagGenesisTime, genesisTime, "")
	c.Flags().String(flagChainID, chainID, "")

	return c
}

func TestGetMigrationCallback(t *testing.T) {
	for _, version := range GetMigrationVersions() {
		require.NotNil(t, GetMigrationCallback(version))
	}
}

func TestMigrateGenesis(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	tests.CreateConfigFolder(t, home)

	logger := log.NewNopLogger()
	cfg := config.TestConfig()
	ctx := server.NewContext(viper.New(), cfg, logger)
	cdc := makeCodec()

	genesisPath := path.Join(home, "genesis.json")
	target := "v0.36"

	// Reject if we dont' have the right parameters or genesis does not exists
	cmd := MigrateGenesisCmd(ctx, cdc)
	cmd.SetArgs([]string{target, genesisPath})
	require.Error(t, cmd.Execute())

	// Noop migration with minimal genesis
	emptyGenesis := []byte(`{"chain_id":"test","app_state":{}}`)
	err := ioutil.WriteFile(genesisPath, emptyGenesis, 0644)
	require.Nil(t, err)
	setupCmd := setupCmd("", "test2")
	require.NoError(t, MigrateGenesisCmd(ctx, cdc).RunE(setupCmd, []string{target, genesisPath}))
	// Every migration function shuold tests its own module separately
}
