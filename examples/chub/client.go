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

	flagBind = "bind"
	flagCORS = "cors"
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

// ClientCommands returns a sub-tree of all basic client commands
//
// Call AddGetCommand and AddPostCommand to add custom txs and queries
func ClientCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Interact with the chain via a light-client",
	}
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
	return cmd
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
	cmd.Flags().StringP(flagNode, "n", "tcp://localhost:46657", "node to connect to")
	cmd.Flags().String(flagGenesis, "", "genesis file to verify header validity")
	cmd.Flags().String(flagCommit, "", "file with trusted and signed header")
	cmd.Flags().String(flagValHash, "", "hash of trusted validator set (hex-encoded)")
	return cmd
}

func blockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block <height>",
		Short: "Get verified data for a the block at given height",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringSlice(flagSelect, []string{"header", "tx"}, "fields to return (header|txs|results)")
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
	cmd.Flags().StringP(flagBind, "b", "localhost:1317", "interface and port that server binds to")
	cmd.Flags().String(flagCORS, "", "Set to domains that can make CORS requests (* for all)")
	return cmd
}

func txSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringSlice(flagTags, nil, "tags that must match (may provide multiple)")
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
