package rpc

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// XXX: remove this when not needed
func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("TODO: Command not yet implemented")
}

const (
	// one of the following should be provided to verify the connection
	flagGenesis = "genesis"
	flagCommit  = "commit"
	flagValHash = "validator-set"

	flagSelect = "select"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  todoNotImplemented,
	}
)

// AddCommands adds a number of rpc-related subcommands
func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		initClientCommand(),
		statusCmd,
		blockCommand(),
		validatorCommand(),
	)
}

func initClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize light client",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringP(client.FlagChainID, "c", "", "ID of chain we connect to")
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
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
