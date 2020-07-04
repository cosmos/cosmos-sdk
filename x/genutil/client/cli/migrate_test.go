package cli

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
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
	viper.Set(cli.HomeFlag, home)
	viper.Set(flags.FlagName, "moniker")
	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := server.NewContext(viper.New(), cfg, logger)
	cdc := makeCodec()

	genesisPath := path.Join(home, "genesis.json")
	target := "v0.36"

	// Reject if we dont' have the right parameters or genesis does not exists
	require.Error(t, MigrateGenesisCmd(ctx, cdc).RunE(nil, []string{target, genesisPath}))

	// Noop migration with minimal genesis
	emptyGenesis := []byte(`{"chain_id":"test","app_state":{}}`)
	err = ioutil.WriteFile(genesisPath, emptyGenesis, 0644)
	require.Nil(t, err)
	cmd := setupCmd("", "test2")
	require.NoError(t, MigrateGenesisCmd(ctx, cdc).RunE(cmd, []string{target, genesisPath}))
	// Every migration function shuold tests its own module separately
}
