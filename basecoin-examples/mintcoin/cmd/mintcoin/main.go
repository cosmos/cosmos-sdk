package main

import (
	"os"

	// import _ to register the mint plugin to apptx
	_ "github.com/tendermint/basecoin-examples/mintcoin/commands"
	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "mintcoin"
	app.Usage = "mintcoin [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.AccountCmd,
	}
	app.Run(os.Args)
}
