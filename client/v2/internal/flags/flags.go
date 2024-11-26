package flags

// This defines flag names that can be used in autocli.
const (
	// FlagHome is the flag to specify the home dir of the app.
	FlagHome = "home"

	// FlagChainID is the flag to specify the chain ID of the network.
	FlagChainID = "chain-id"

	// FlagFrom is the flag to set the from address with which to sign the transaction.
	FlagFrom = "from"

	// FlagOutput is the flag to set the output format.
	FlagOutput = "output"

	// FlagNoIndent is the flag to not indent the output.
	FlagNoIndent = "no-indent"

	// FlagNoPrompt is the flag to not use a prompt for commands.
	FlagNoPrompt = "no-prompt"

	// FlagKeyringDir is the flag to specify the directory where the keyring is stored.
	FlagKeyringDir = "keyring-dir"
	// FlagKeyringBackend is the flag to specify which backend to use for the keyring (e.g. os, file, test).
	FlagKeyringBackend = "keyring-backend"

	// FlagNoProposal is the flag convert a gov proposal command into a normal command.
	// This is used to allow user of chains with custom authority to not use gov submit proposals for usual proposal commands.
	FlagNoProposal = "no-proposal"

	// FlagNode is the flag to specify the node address to connect to.
	FlagNode = "node"
	// FlagBroadcastMode is the flag to specify the broadcast mode for transactions.
	FlagBroadcastMode = "broadcast-mode"

	// FlagGrpcAddress is the flag to specify the gRPC server address to connect to.
	FlagGrpcAddress = "grpc-addr"
	// FlagGrpcInsecure is the flag to allow insecure gRPC connections.
	FlagGrpcInsecure = "grpc-insecure"
)

// List of supported output formats
const (
	OutputFormatJSON = "json"
	OutputFormatText = "text"
)
