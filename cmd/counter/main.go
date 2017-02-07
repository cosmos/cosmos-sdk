package main

import (
	"os"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "counter"
	app.Usage = "counter [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.TxCmd,
		commands.KeyCmd,
		commands.QueryCmd,
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
