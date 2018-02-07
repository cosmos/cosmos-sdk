package main

import "github.com/spf13/cobra"

const (
	// these are needed for every init
	flagChainID = "chain-id"
	flagNode    = "node"

	// one of the following should be provided to verify the connection
	flagGenesis = "genesis"
	flagCommit  = "commit"
	flagValHash = "validator-set"

	flagSelect = "select"
	flagTags   = "tag"
	flagAny    = "any"

	flagBind      = "bind"
	flagCORS      = "cors"
	flagTrustNode = "trust-node"

	// this is for signing
	flagName = "name"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  todoNotImplemented,
	}

	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Query ABCI state data",
		Long: `Query ABCI state data.
Subcommands should be defined for each particular object to query.`,
	}

	postCmd = &cobra.Command{
		Use:   "post",
		Short: "Create a new transaction and post it to the chain",
		Long: `Create a new transaction and post it to the chain.
Subcommands should be defined for each particular transaction type.`,
	}
)

func init() {
	getCmd.PersistentFlags().Bool(flagTrustNode, false, "Don't verify proofs for responses")

	postCmd.PersistentFlags().String(flagName, "", "Name of private key with which to sign")
	postCmd.PersistentFlags().String(flagPassword, "", "Password to use the named private key")
}

// AddClientCommands returns a sub-tree of all basic client commands
//
// Call AddGetCommand and AddPostCommand to add custom txs and queries
func AddClientCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		initClientCommand(),
		statusCmd,
		blockCommand(),
		validatorCommand(),
		lineBreak,
		txSearchCommand(),
		txCommand(),
		lineBreak,
		getCmd,
		postCmd,
		lineBreak,
		serveCommand(),
	)
}

// AddGetCommand adds one or more query subcommands
func AddGetCommand(cmds ...*cobra.Command) {
	getCmd.AddCommand(cmds...)
}

// AddPostCommand adds one or more subcommands to create transactions
func AddPostCommand(cmds ...*cobra.Command) {
	postCmd.AddCommand(cmds...)
}

func initClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize light client",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringP(flagChainID, "c", "", "ID of chain we connect to")
	cmd.Flags().StringP(flagNode, "n", "tcp://localhost:46657", "Node to connect to")
	cmd.Flags().String(flagGenesis, "", "Genesis file to verify header validity")
	cmd.Flags().String(flagCommit, "", "File with trusted and signed header")
	cmd.Flags().String(flagValHash, "", "Hash of trusted validator set (hex-encoded)")
	return cmd
}

func blockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block <height>",
		Short: "Get verified data for a the block at given height",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringSlice(flagSelect, []string{"header", "tx"}, "Fields to return (header|txs|results)")
	return cmd
}

func validatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validatorset <height>",
		Short: "Get the full validator set at given height",
		RunE:  todoNotImplemented,
	}
	return cmd
}

func serveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE:  todoNotImplemented,
	}
	// TODO: handle unix sockets also?
	cmd.Flags().StringP(flagBind, "b", "localhost:1317", "Interface and port that server binds to")
	cmd.Flags().String(flagCORS, "", "Set to domains that can make CORS requests (* for all)")
	return cmd
}

func txSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringSlice(flagTags, nil, "Tags that must match (may provide multiple)")
	cmd.Flags().Bool(flagAny, false, "Return transactions that match ANY tag, rather than ALL")
	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx <hash>",
		Short: "Matches this txhash over all committed blocks",
		RunE:  todoNotImplemented,
	}
	return cmd
}
