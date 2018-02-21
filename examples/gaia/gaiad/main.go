package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

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
func defaultOptions(args []string) (json.RawMessage, error) {
	addr, secret, err := server.GenerateCoinKey()
	if err != nil {
		return nil, err
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
	return json.RawMessage(opts), nil
}

func main() {
	// TODO: set this to something real
	app := new(baseapp.BaseApp)

	gaiadCmd.AddCommand(
		server.InitCmd(defaultOptions, app.Logger),
		server.StartCmd(app, app.Logger),
		server.UnsafeResetAllCmd(app.Logger),
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(gaiadCmd, "GA", os.ExpandEnv("$HOME/.gaiad"))
	executor.Execute()
}
