package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/abci/types"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tmlibs/cli"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/version"
)

// gaiadCmd is the entry point for this binary
var (
	context  = server.NewContext(nil, nil)
	gaiadCmd = &cobra.Command{
		Use:   "gaiad",
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
      }]
    }`, addr)
	return json.RawMessage(opts), secret, addr, nil
}

func generateApp(rootDir string, logger log.Logger) (abci.Application, error) {
	// TODO: set this to something real
	app := new(baseapp.BaseApp)
	return app, nil
}

func main() {
	gaiadCmd.AddCommand(
		server.InitCmd(defaultOptions, context),
		server.StartCmd(generateApp, context),
		server.UnsafeResetAllCmd(context),
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(gaiadCmd, "GA", os.ExpandEnv("$HOME/.gaiad"))
	executor.Execute()
}
