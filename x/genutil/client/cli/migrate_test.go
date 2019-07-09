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

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
)

func setupCmd(genesisTime string, chainId string) *cobra.Command {
	c := &cobra.Command{
		Use:  "c",
		Args: cobra.ArbitraryArgs,
		Run:  func(_ *cobra.Command, args []string) {},
	}

	c.Flags().String(flagGenesisTime, genesisTime, "")
	c.Flags().String(flagChainId, chainId, "")

	return c
}

func TestMigrateGenesis(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	viper.Set(cli.HomeFlag, home)
	viper.Set(client.FlagName, "moniker")
	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := server.NewContext(cfg, logger)
	cdc := makeCodec()

	genesisPath := path.Join(home, "genesis.json")
	target := "v0.36"

	defer cleanup()

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
