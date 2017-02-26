package main

import (
	"os"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "basecoin"
	app.Usage = "basecoin [command] [args...]"
	app.Version = version.Version
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.KeyCmd,
		commands.VerifyCmd,
		commands.BlockCmd,
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
