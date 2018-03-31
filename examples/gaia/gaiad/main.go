package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/version"
)

// gaiadCmd is the entry point for this binary
var (
	gaiadCmd = &cobra.Command{
		Use:   "gaiad",
		Short: "Gaia Daemon (server)",
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
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "main")

	gaiadCmd.AddCommand(
		server.InitCmd(defaultOptions, logger),
		server.StartCmd(generateApp, logger),
		server.UnsafeResetAllCmd(logger),
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(gaiadCmd, "GA", os.ExpandEnv("$HOME/.gaiad"))
	executor.Execute()
}
