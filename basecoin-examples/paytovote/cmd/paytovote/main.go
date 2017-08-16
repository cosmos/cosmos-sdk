package main

import (
	"os"

	_ "github.com/tendermint/basecoin-examples/paytovote/commands"
	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "paytovote"
	app.Usage = "paytovote [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
