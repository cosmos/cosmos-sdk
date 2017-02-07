package main

import (
	"os"

	"github.com/tendermint/basecoin/cmd/commands"

	"github.com/urfave/cli"
)

func init() {
	commands.RegisterIBC()
}

func main() {
	app := cli.NewApp()
	app.Name = "adam"
	app.Usage = "adam [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.VerifyCmd, // TODO: move to merkleeyes?
		commands.BlockCmd,
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
