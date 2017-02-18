package main

import (
	"os"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/urfave/cli"
)

func main() {
	//Initialize an instance of basecoin with default basecoin commands
	app := cli.NewApp()
	app.Name = "example-plugin"
	app.Usage = "example-plugin [command] [args...]"
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
