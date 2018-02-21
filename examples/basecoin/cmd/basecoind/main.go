package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
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
	// TODO: this should somehow be updated on cli flags?
	// But we need to create the app first... hmmm.....
	rootDir := os.ExpandEnv("$HOME/.basecoind")

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	db, err := dbm.NewGoLevelDB("basecoin", rootDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	bapp := app.NewBasecoinApp(logger, db)

	gaiadCmd.AddCommand(
		server.InitCmd(defaultOptions, bapp.Logger),
		server.StartCmd(bapp, bapp.Logger),
		server.UnsafeResetAllCmd(bapp.Logger),
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(gaiadCmd, "BC", rootDir)
	executor.Execute()
}
