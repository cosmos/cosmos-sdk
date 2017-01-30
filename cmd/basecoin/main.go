package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "basecoin"
	app.Usage = "basecoin [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		startCmd,
		sendTxCmd,
		appTxCmd,
		ibcCmd,
		queryCmd,
		blockCmd,
		accountCmd,
	}
	app.Run(os.Args)
}
