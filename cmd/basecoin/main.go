package main

import (
	"os"

	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "basecoin"
	app.Usage = "basecoin [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.SendTxCmd,
		commands.AppTxCmd,
		commands.IbcCmd,
		commands.QueryCmd,
		commands.VerifyCmd,
		commands.BlockCmd,
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
