package main

import (
	"os"

	// import _ to register the stake plugin to apptx
	_ "github.com/tendermint/basecoin-examples/stake/commands"
	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "stakecoin"
	app.Usage = "stakecoin [command] [args...]"
	app.Version = "0.0.0"
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
