package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/abci/types"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tmlibs/cli"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/democoin/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/version"
)

// democoindCmd is the entry point for this binary
var (
	context      = server.NewContext(nil, nil)
	democoindCmd = &cobra.Command{
		Use:   "democoind",
		Short: "Gaia Daemon (server)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == version.VersionCmd.Name() {
				return nil
			}
			config, err := tcmd.ParseConfig()
			if err != nil {
				return err
			}
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
			logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
			if err != nil {
				return err
			}
			if viper.GetBool(cli.TraceFlag) {
				logger = log.NewTracingLogger(logger)
			}
			logger = logger.With("module", "main")
			context.Config = config
			context.Logger = logger
			return nil
		},
	}
)

// defaultOptions sets up the app_options for the
// default genesis file
func defaultOptions(args []string) (json.RawMessage, string, cmn.HexBytes, error) {
	addr, secret, err := server.GenerateCoinKey()
	if err != nil {
		return nil, "", nil, err
	}
	fmt.Println("Secret phrase to access coins:")
	fmt.Println(secret)

	opts := fmt.Sprintf(`{
      "accounts": [{
        "address": "%s",
        "coins": [
          {
            "denom": "mycoin",
            "amount": 9007199254740992
          }
        ]
      }],
      "cool": {
        "trend": "ice-cold"
      }
    }`, addr)
	return json.RawMessage(opts), "", nil, nil
}

func generateApp(rootDir string, logger log.Logger) (abci.Application, error) {
	dbMain, err := dbm.NewGoLevelDB("democoin", filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, err
	}
	dbAcc, err := dbm.NewGoLevelDB("democoin-acc", filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, err
	}
	dbIBC, err := dbm.NewGoLevelDB("democoin-ibc", filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, err
	}
	dbStaking, err := dbm.NewGoLevelDB("democoin-staking", filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, err
	}
	dbs := map[string]dbm.DB{
		"main":    dbMain,
		"acc":     dbAcc,
		"ibc":     dbIBC,
		"staking": dbStaking,
	}
	bapp := app.NewDemocoinApp(logger, dbs)
	return bapp, nil
}

func main() {
	democoindCmd.AddCommand(
		server.InitCmd(defaultOptions, context),
		server.StartCmd(generateApp, context),
		server.UnsafeResetAllCmd(context),
		server.ShowNodeIdCmd(context),
		server.ShowValidatorCmd(context),
		version.VersionCmd,
	)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.democoind")
	executor := cli.PrepareBaseCmd(democoindCmd, "BC", rootDir)
	executor.Execute()
}
