package rpc

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
)

const (
	// one of the following should be provided to verify the connection
	flagGenesis = "genesis"
	flagCommit  = "commit"
	flagValHash = "validator-set"
)

// XXX: remove this when not needed
func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("todo: Command not yet implemented")
}

// AddCommands adds a number of rpc-related subcommands
func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		initClientCommand(),
		statusCommand(),
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

// Register REST endpoints
func RegisterRoutes(ctx context.CoreContext, r *mux.Router) {
	r.HandleFunc("/node_info", NodeInfoRequestHandlerFn(ctx)).Methods("GET")
	r.HandleFunc("/syncing", NodeSyncingRequestHandlerFn(ctx)).Methods("GET")
	r.HandleFunc("/blocks/latest", LatestBlockRequestHandlerFn(ctx)).Methods("GET")
	r.HandleFunc("/blocks/{height}", BlockRequestHandlerFn(ctx)).Methods("GET")
	r.HandleFunc("/validatorsets/latest", LatestValidatorSetRequestHandlerFn(ctx)).Methods("GET")
	r.HandleFunc("/validatorsets/{height}", ValidatorSetRequestHandlerFn(ctx)).Methods("GET")
}
