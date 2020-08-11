package cli_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

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

// custom tx codec
func makeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

func setupCmd(genesisTime string, chainID string) *cobra.Command {
	c := &cobra.Command{
		Use:  "c",
		Args: cobra.ArbitraryArgs,
		Run:  func(_ *cobra.Command, args []string) {},
	}

	c.Flags().String("genesis-time", genesisTime, "")
	c.Flags().String("chain-id", chainID, "")

	return c
}

func TestGetMigrationCallback(t *testing.T) {
	for _, version := range genutilcli.GetMigrationVersions() {
		require.NotNil(t, genutilcli.GetMigrationCallback(version))
	}
}

func TestMigrateGenesis(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	viper.Set(cli.HomeFlag, home)
	viper.Set(flags.FlagName, "moniker")
	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := server.NewContext(cfg, logger)
	cdc := makeCodec()

	genesisPath := path.Join(home, "genesis.json")
	target := "v0.36"

	defer cleanup()

	// Reject if we dont' have the right parameters or genesis does not exists
	require.Error(t, genutilcli.MigrateGenesisCmd(ctx, cdc).RunE(nil, []string{target, genesisPath}))

	// Noop migration with minimal genesis
	emptyGenesis := []byte(`{"chain_id":"test","app_state":{}}`)
	err = ioutil.WriteFile(genesisPath, emptyGenesis, 0600)
	require.Nil(t, err)
	cmd := setupCmd("", "test2")
	require.NoError(t, genutilcli.MigrateGenesisCmd(ctx, cdc).RunE(cmd, []string{target, genesisPath}))
	// Every migration function shuold tests its own module separately
}

func TestMigrateCommercioGenesisData(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	viper.Set(cli.HomeFlag, home)
	viper.Set(flags.FlagName, "moniker")
	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := server.NewContext(cfg, logger)
	cdc := makeCodec()

	genesisPath := filepath.Join("testdata", "commercio-genesis.json")
	require.NoError(t, err)
	target := "v0.39"

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("did:com:", "did:com:pub")
	config.Seal()

	// Migration with minimal genesis
	cmd := setupCmd("2020-08-10T09:52:06.576222474Z", "test2")
	_, newGenesisStream, _ := tests.ApplyMockIO(cmd)
	require.NoError(t, genutilcli.MigrateGenesisCmd(ctx, cdc).RunE(cmd, []string{target, genesisPath}))

	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, 0)
	newGenesisBytes := newGenesisStream.Bytes()

	// Initialize the chain
	require.NotPanics(t, func() {
		app.InitChain(
			abci.RequestInitChain{
				Validators:    []abci.ValidatorUpdate{},
				AppStateBytes: newGenesisBytes,
			},
		)
	})
}
