package rpc

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

const (
	// one of the following should be provided to verify the connection
	flagGenesis = "genesis"
	flagCommit  = "commit"
	flagValHash = "validator-set"
)

// XXX: remove this when not needed
func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("TODO: Command not yet implemented")
}

// AddCommands adds a number of rpc-related subcommands
func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		initClientCommand(),
		statusCommand(),
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

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/node_info", NodeInfoRequestHandler).Methods("GET")
	r.HandleFunc("/syncing", NodeSyncingRequestHandler).Methods("GET")
	r.HandleFunc("/blocks/latest", LatestBlockRequestHandler).Methods("GET")
	r.HandleFunc("/blocks/{height}", BlockRequestHandler).Methods("GET")
	r.HandleFunc("/validatorsets/latest", LatestValidatorsetRequestHandler).Methods("GET")
	r.HandleFunc("/validatorsets/{height}", ValidatorsetRequestHandler).Methods("GET")
}
