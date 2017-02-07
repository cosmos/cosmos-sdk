package main

import (
	"os"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "basecoin"
	app.Usage = "basecoin [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.KeyCmd,
		commands.VerifyCmd, // TODO: move to merkleeyes?
		commands.BlockCmd,  // TODO: move to adam?
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
