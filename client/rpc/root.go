package rpc

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/spf13/viper"
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

// InitClientCommand initializes client commands
func InitClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize light client",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringP(client.FlagChainID, "c", "", "ID of chain we connect to")
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String(flagGenesis, "", "Genesis file to verify header validity")
	cmd.Flags().String(flagCommit, "", "File with trusted and signed header")
	cmd.Flags().String(flagValHash, "", "Hash of trusted validator set (hex-encoded)")
	viper.BindPFlag(client.FlagChainID, cmd.Flags().Lookup(client.FlagChainID))
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))

	return cmd
}

// Register REST endpoints
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/node_info", NodeInfoRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/syncing", NodeSyncingRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/blocks/latest", LatestBlockRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/blocks/{height}", BlockRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/validatorsets/latest", LatestValidatorSetRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/validatorsets/{height}", ValidatorSetRequestHandlerFn(cliCtx)).Methods("GET")
}
