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
			dirFlag,
			inProcTMFlag,
			chainIDFlag,
			ibcPluginFlag,
			counterPluginFlag,
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
			nodeFlag,
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
			nodeFlag,
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

	ibcCmd = cli.Command{
		Name:  "ibc",
		Usage: "Send a transaction to the interblockchain (ibc) plugin",
		Flags: []cli.Flag{
			nodeFlag,
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
			ibcRegisterTxCmd,
			ibcUpdateTxCmd,
			ibcPacketTxCmd,
		},
	}

	ibcRegisterTxCmd = cli.Command{
		Name:  "register",
		Usage: "Register a blockchain via IBC",
		Action: func(c *cli.Context) error {
			return cmdIBCRegisterTx(c)
		},
		Flags: []cli.Flag{
			ibcChainIDFlag,
			ibcGenesisFlag,
		},
	}

	ibcUpdateTxCmd = cli.Command{
		Name:  "update",
		Usage: "Update the latest state of a blockchain via IBC",
		Action: func(c *cli.Context) error {
			return cmdIBCUpdateTx(c)
		},
		Flags: []cli.Flag{
			ibcHeaderFlag,
			ibcCommitFlag,
		},
	}

	ibcPacketTxCmd = cli.Command{
		Name:  "packet",
		Usage: "Send a new packet via IBC",
		Flags: []cli.Flag{
		//
		},
		Subcommands: []cli.Command{
			ibcPacketCreateTx,
			ibcPacketPostTx,
		},
	}

	ibcPacketCreateTx = cli.Command{
		Name:  "create",
		Usage: "Create an egress IBC packet",
		Action: func(c *cli.Context) error {
			return cmdIBCPacketCreateTx(c)
		},
		Flags: []cli.Flag{
			ibcFromFlag,
			ibcToFlag,
			ibcTypeFlag,
			ibcPayloadFlag,
			ibcSequenceFlag,
		},
	}

	ibcPacketPostTx = cli.Command{
		Name:  "post",
		Usage: "Deliver an IBC packet to another chain",
		Action: func(c *cli.Context) error {
			return cmdIBCPacketPostTx(c)
		},
		Flags: []cli.Flag{
			ibcFromFlag,
			ibcHeightFlag,
			ibcPacketFlag,
			ibcProofFlag,
		},
	}

	queryCmd = cli.Command{
		Name:      "query",
		Usage:     "Query the merkle tree",
		ArgsUsage: "<key>",
		Action: func(c *cli.Context) error {
			return cmdQuery(c)
		},
		Flags: []cli.Flag{
			nodeFlag,
		},
	}

	accountCmd = cli.Command{
		Name:      "account",
		Usage:     "Get details of an account",
		ArgsUsage: "<address>",
		Action: func(c *cli.Context) error {
			return cmdAccount(c)
		},
		Flags: []cli.Flag{
			nodeFlag,
		},
	}

	blockCmd = cli.Command{
		Name:      "block",
		Usage:     "Get the header and commit of a block",
		ArgsUsage: "<height>",
		Action: func(c *cli.Context) error {
			return cmdBlock(c)
		},
		Flags: []cli.Flag{
			nodeFlag,
		},
	}

	verifyCmd = cli.Command{
		Name:  "verify",
		Usage: "Verify the IAVL proof",
		Action: func(c *cli.Context) error {
			return cmdVerify(c)
		},
		Flags: []cli.Flag{
			proofFlag,
			keyFlag,
			valueFlag,
			rootFlag,
		},
	}
)
