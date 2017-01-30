package commands

import (
	"github.com/urfave/cli"
)

// start flags
var (
	AddrFlag = cli.StringFlag{
		Name:  "address",
		Value: "tcp://0.0.0.0:46658",
		Usage: "Listen address",
	}

	EyesFlag = cli.StringFlag{
		Name:  "eyes",
		Value: "local",
		Usage: "MerkleEyes address, or 'local' for embedded",
	}

	// TODO: move to config file
	// eyesCacheSizePtr := flag.Int("eyes-cache-size", 10000, "MerkleEyes db cache size, for embedded")

	DirFlag = cli.StringFlag{
		Name:  "dir",
		Value: ".",
		Usage: "Root directory",
	}

	InProcTMFlag = cli.BoolFlag{
		Name:  "in-proc",
		Usage: "Run Tendermint in-process with the App",
	}

	IbcPluginFlag = cli.BoolFlag{
		Name:  "ibc-plugin",
		Usage: "Enable the ibc plugin",
	}
)

// tx flags

var (
	NodeFlag = cli.StringFlag{
		Name:  "node",
		Value: "tcp://localhost:46657",
		Usage: "Tendermint RPC address",
	}

	ToFlag = cli.StringFlag{
		Name:  "to",
		Value: "",
		Usage: "Destination address for the transaction",
	}

	AmountFlag = cli.IntFlag{
		Name:  "amount",
		Value: 0,
		Usage: "Amount of coins to send in the transaction",
	}

	FromFlag = cli.StringFlag{
		Name:  "from",
		Value: "priv_validator.json",
		Usage: "Path to a private key to sign the transaction",
	}

	SeqFlag = cli.IntFlag{
		Name:  "sequence",
		Value: 0,
		Usage: "Sequence number for the account",
	}

	CoinFlag = cli.StringFlag{
		Name:  "coin",
		Value: "blank",
		Usage: "Specify a coin denomination",
	}

	GasFlag = cli.IntFlag{
		Name:  "gas",
		Value: 0,
		Usage: "The amount of gas for the transaction",
	}

	FeeFlag = cli.IntFlag{
		Name:  "fee",
		Value: 0,
		Usage: "The transaction fee",
	}

	DataFlag = cli.StringFlag{
		Name:  "data",
		Value: "",
		Usage: "Data to send with the transaction",
	}

	NameFlag = cli.StringFlag{
		Name:  "name",
		Value: "",
		Usage: "Plugin to send the transaction to",
	}

	ChainIDFlag = cli.StringFlag{
		Name:  "chain_id",
		Value: "test_chain_id",
		Usage: "ID of the chain for replay protection",
	}

	ValidFlag = cli.BoolFlag{
		Name:  "valid",
		Usage: "Set valid field in CounterTx",
	}
)

// ibc flags
var (
	IbcChainIDFlag = cli.StringFlag{
		Name:  "chain_id",
		Usage: "ChainID for the new blockchain",
		Value: "",
	}

	IbcGenesisFlag = cli.StringFlag{
		Name:  "genesis",
		Usage: "Genesis file for the new blockchain",
		Value: "",
	}

	IbcHeaderFlag = cli.StringFlag{
		Name:  "header",
		Usage: "Block header for an ibc update",
		Value: "",
	}

	IbcCommitFlag = cli.StringFlag{
		Name:  "commit",
		Usage: "Block commit for an ibc update",
		Value: "",
	}

	IbcFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "Source ChainID",
		Value: "",
	}

	IbcToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "Destination ChainID",
		Value: "",
	}

	IbcTypeFlag = cli.StringFlag{
		Name:  "type",
		Usage: "IBC packet type (eg. coin)",
		Value: "",
	}

	IbcPayloadFlag = cli.StringFlag{
		Name:  "payload",
		Usage: "IBC packet payload",
		Value: "",
	}

	IbcPacketFlag = cli.StringFlag{
		Name:  "packet",
		Usage: "hex-encoded IBC packet",
		Value: "",
	}

	IbcProofFlag = cli.StringFlag{
		Name:  "proof",
		Usage: "hex-encoded proof of IBC packet from source chain",
		Value: "",
	}

	IbcSequenceFlag = cli.IntFlag{
		Name:  "sequence",
		Usage: "sequence number for IBC packet",
		Value: 0,
	}

	IbcHeightFlag = cli.IntFlag{
		Name:  "height",
		Usage: "Height the packet became egress in source chain",
		Value: 0,
	}
)

// proof flags
var (
	ProofFlag = cli.StringFlag{
		Name:  "proof",
		Usage: "hex-encoded IAVL proof",
		Value: "",
	}

	KeyFlag = cli.StringFlag{
		Name:  "key",
		Usage: "key to the IAVL tree",
		Value: "",
	}

	ValueFlag = cli.StringFlag{
		Name:  "value",
		Usage: "value in the IAVL tree",
		Value: "",
	}

	RootFlag = cli.StringFlag{
		Name:  "root",
		Usage: "root hash of the IAVL tree",
		Value: "",
	}
)
