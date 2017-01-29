package main

import (
	"github.com/urfave/cli"
)

var (
	startCmd = cli.Command{
		Name:      "start",
		Usage:     "Start basecoin",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdStart(c)
		},
		Flags: []cli.Flag{
			addrFlag,
			eyesFlag,
			eyesDBFlag,
			genesisFlag,
			inProcTMFlag,
			chainIDFlag,
			pluginFlag,
		},
	}

	sendTxCmd = cli.Command{
		Name:      "sendtx",
		Usage:     "Broadcast a basecoin SendTx",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdSendTx(c)
		},
		Flags: []cli.Flag{
			tmAddrFlag,
			chainIDFlag,

			fromFlag,

			amountFlag,
			coinFlag,
			gasFlag,
			feeFlag,
			seqFlag,

			toFlag,
		},
	}

	appTxCmd = cli.Command{
		Name:      "apptx",
		Usage:     "Broadcast a basecoin AppTx",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdAppTx(c)
		},
		Flags: []cli.Flag{
			tmAddrFlag,
			chainIDFlag,

			fromFlag,

			amountFlag,
			coinFlag,
			gasFlag,
			feeFlag,
			seqFlag,

			nameFlag,
			dataFlag,
		},
		Subcommands: []cli.Command{
			counterTxCmd,
		},
	}

	counterTxCmd = cli.Command{
		Name:  "counter",
		Usage: "Craft a transaction to the counter plugin",
		Action: func(c *cli.Context) error {
			return cmdCounterTx(c)
		},
		Flags: []cli.Flag{
			validFlag,
		},
	}

	accountCmd = cli.Command{
		Name:      "account",
		Usage:     "Get details of an account",
		ArgsUsage: "[address]",
		Action: func(c *cli.Context) error {
			return cmdAccount(c)
		},
		Flags: []cli.Flag{
			tmAddrFlag,
		},
	}
)
