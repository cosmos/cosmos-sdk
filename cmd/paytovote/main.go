package main

import (
	"os"

	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	_ "github.com/tendermint/basecoin/cmd/paytovote/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "paytovote"
	app.Usage = "paytovote [command] [args...]"
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
