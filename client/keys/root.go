package keys

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
)

// Commands registers a sub-tree of commands to interact with
// local private key storage.
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Add or view local private keys",
		Long: `Keys allows you to manage your local keystore for tendermint.

    These keys may be in any format supported by go-crypto and can be
    used by light-clients, full nodes, or any other application that
    needs to sign with a private key.`,
	}
	cmd.AddCommand(
		mnemonicKeyCommand(),
		addKeyCommand(),
		listKeysCmd(),
		showKeysCmd(),
		client.LineBreak,
		deleteKeyCommand(),
		updateKeyCommand(),
	)
	return cmd
}

// register REST routes
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/keys", QueryKeysRequestHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/keys", AddNewKeyRequestHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/keys/seed", SeedRequestHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/keys/{name}/recover", RecoverRequestHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/keys/{name}", GetKeyRequestHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/keys/{name}", UpdateKeyRequestHandler(cliCtx)).Methods("PUT")
	r.HandleFunc("/keys/{name}", DeleteKeyRequestHandler(cliCtx)).Methods("DELETE")
}
