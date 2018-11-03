package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/p2p"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/examples/democoin/app"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	flagClientHome = "home-client"
)

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

			// XXX gaiaInit -> server
			pk := gaiaInit.ReadOrCreatePrivValidator(config.PrivValidatorFile())
			validator := tmtypes.GenesisValidator{
				PubKey: pk,
				Power:  10,
			}

			addr, secret, err := server.GenerateAccountKeyAndSecret()
			if err != nil {
				return err
			}
			appStateJSON := (fmt.Sprintf(`{
  "cool": {
    "trend": "ice-cold"
  },
  "pow": {
    "difficulty": "1",
    "count": "0"
  },
  "accounts": [{
    "address": "%s",
    "coins": [
      {
        "denom": "mycoin",
        "amount": "9007199254740992"
      }
    ]
  }]
}`, addr))

			toPrint := struct {
				ChainID string `json:"chain_id"`
				NodeID  string `json:"node_id"`
				Message string `json:"message"`
			}{
				chainID,
				nodeID,
				fmt.Sprintf("Secret: %v", secret),
			}
			out, err := codec.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "%s\n", string(out))
			return gaiaInit.WriteGenesisFile(config.GenesisFile(), chainID,
				[]tmtypes.GenesisValidator{validator}, []byte(appStateJSON))
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

func newApp(logger log.Logger, db dbm.DB, _ io.Writer) abci.Application {
	return app.NewDemocoinApp(logger, db)
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB, _ io.Writer) (
	json.RawMessage, []tmtypes.GenesisValidator, error) {
	dapp := app.NewDemocoinApp(logger, db)
	return dapp.ExportAppStateAndValidators()
}

func main() {
	cdc := app.MakeCodec()

	// Setup certain SDK config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("demoacc", "demopub")
	config.SetBech32PrefixForValidator("demoval", "demovalpub")
	config.SetBech32PrefixForConsensusNode("democons", "democonspub")
	config.Seal()

	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "democoind",
		Short:             "Democoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(InitCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.TestnetFilesCmd(ctx, cdc))

	server.AddCommands(ctx, cdc, rootCmd,
		newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.democoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}
