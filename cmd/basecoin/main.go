package main

import (
	"os"

	"github.com/urfave/cli"
)

// start flags
var (
	addrFlag = cli.StringFlag{
		Name:  "address",
		Value: "tcp://0.0.0.0:46658",
		Usage: "Listen address",
	}

	eyesFlag = cli.StringFlag{
		Name:  "eyes",
		Value: "local",
		Usage: "MerkleEyes address, or 'local' for embedded",
	}

	eyesDBFlag = cli.StringFlag{
		Name:  "eyes-db",
		Value: "merkleeyes.db",
		Usage: "MerkleEyes db name for embedded",
	}

	// TODO: move to config file
	// eyesCacheSizePtr := flag.Int("eyes-cache-size", 10000, "MerkleEyes db cache size, for embedded")

	genesisFlag = cli.StringFlag{
		Name:  "genesis",
		Value: "",
		Usage: "Path to genesis file, if it exists",
	}

	inProcTMFlag = cli.BoolFlag{
		Name:  "in-proc",
		Usage: "Run Tendermint in-process with the App",
	}
)

// tx flags

var (
	tmAddrFlag = cli.StringFlag{
		Name:  "tendermint",
		Value: "tcp://localhost:46657",
		Usage: "Tendermint RPC address",
	}

	toFlag = cli.StringFlag{
		Name:  "to",
		Value: "",
		Usage: "Destination address for the transaction",
	}

	amountFlag = cli.IntFlag{
		Name:  "amount",
		Value: 0,
		Usage: "Amount of coins to send in the transaction",
	}

	fromFlag = cli.StringFlag{
		Name:  "from",
		Value: "priv_validator.json",
		Usage: "Path to a private key to sign the transaction",
	}

	seqFlag = cli.IntFlag{
		Name:  "sequence",
		Value: 0,
		Usage: "Sequence number for the account",
	}

	coinFlag = cli.StringFlag{
		Name:  "coin",
		Value: "blank",
		Usage: "Specify a coin denomination",
	}

	gasFlag = cli.IntFlag{
		Name:  "gas",
		Value: 0,
		Usage: "The amount of gas for the transaction",
	}

	feeFlag = cli.IntFlag{
		Name:  "fee",
		Value: 0,
		Usage: "The transaction fee",
	}

	dataFlag = cli.StringFlag{
		Name:  "data",
		Value: "",
		Usage: "Data to send with the transaction",
	}

	nameFlag = cli.StringFlag{
		Name:  "name",
		Value: "",
		Usage: "Plugin to send the transaction to",
	}

	chainIDFlag = cli.StringFlag{
		Name:  "chain_id",
		Value: "test_chain_id",
		Usage: "ID of the chain for replay protection",
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "basecoin"
	app.Usage = "basecoin [command] [args...]"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
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
			},
		},

		{
			Name:      "sendtx",
			Usage:     "Broadcast a basecoin SendTx",
			ArgsUsage: "",
			Action: func(c *cli.Context) error {
				return cmdSendTx(c)
			},
			Flags: []cli.Flag{
				tmAddrFlag,
				toFlag,
				fromFlag,
				amountFlag,
				coinFlag,
				gasFlag,
				feeFlag,
				chainIDFlag,
				seqFlag,
			},
		},

		{
			Name:      "apptx",
			Usage:     "Broadcast a basecoin AppTx",
			ArgsUsage: "",
			Action: func(c *cli.Context) error {
				return cmdAppTx(c)
			},
			Flags: []cli.Flag{
				tmAddrFlag,
				nameFlag,
				fromFlag,
				amountFlag,
				coinFlag,
				gasFlag,
				feeFlag,
				dataFlag,
				seqFlag,
			},
		},

		{
			Name:      "account",
			Usage:     "Get details of an account",
			ArgsUsage: "",
			Action: func(c *cli.Context) error {
				return cmdAccount(c)
			},
			Flags: []cli.Flag{
				tmAddrFlag,
			},
		},
	}
	app.Run(os.Args)
}
