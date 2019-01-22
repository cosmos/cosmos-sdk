package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/store"

	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"

	"github.com/cosmos/cosmos-sdk/baseapp"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/docs/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	flagClientHome = "home-client"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "basecoind",
		Short:             "Basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(InitCmd(ctx, cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.basecoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)

	// initialise the Bech32 prefixes
	initSDKConfig()

	err := executor.Execute()
	if err != nil {
		// Note: Handle with #870
		panic(err)
	}
}

func initSDKConfig() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("baseacc", "basepub")
	config.SetBech32PrefixForValidator("baseval", "basevalpub")
	config.SetBech32PrefixForConsensusNode("basecons", "baseconspub")
	config.Seal()
}

// get cmd to initialize all files for tendermint and application
// nolint: errcheck
func InitCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis config, priv-validator file, and p2p-node file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))
			chainID := viper.GetString(client.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
			if err != nil {
				return err
			}
			nodeID := string(nodeKey.ID())

			pk := privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(),
				config.PrivValidatorStateFile()).GetPubKey()
			genTx, appMessage, validator, err := server.SimpleAppGenTx(cdc, pk)
			if err != nil {
				return err
			}

			appState, err := server.SimpleAppGenState(
				cdc, tmtypes.GenesisDoc{}, []json.RawMessage{genTx})
			if err != nil {
				return err
			}
			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			toPrint := struct {
				ChainID    string          `json:"chain_id"`
				NodeID     string          `json:"node_id"`
				AppMessage json.RawMessage `json:"app_message"`
			}{
				chainID,
				nodeID,
				appMessage,
			}
			out, err := codec.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "%s\n", string(out))
			return gaiaInit.ExportGenesisFile(config.GenesisFile(), chainID,
				[]tmtypes.GenesisValidator{validator}, appStateJSON)
		},
	}

	cmd.Flags().String(cli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, app.DefaultCLIHome, "client's home directory")
	cmd.Flags().String(client.FlagChainID, "",
		"genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(client.FlagName, "", "validator's moniker")
	cmd.MarkFlagRequired(client.FlagName)
	return cmd
}

func newApp(logger log.Logger, db dbm.DB, storeTracer io.Writer) abci.Application {
	return app.NewBasecoinApp(logger, db, baseapp.SetPruning(store.NewPruningOptions(viper.GetString("pruning"))))
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB, storeTracer io.Writer, _ int64, _ bool) (
	json.RawMessage, []tmtypes.GenesisValidator, error) {
	bapp := app.NewBasecoinApp(logger, db)
	return bapp.ExportAppStateAndValidators()
}
