package keys

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
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
		newKeyCommand(),
		addKeyCommand(),
		listKeysCmd,
		showKeysCmd(),
		client.LineBreak,
		deleteKeyCommand(),
		updateKeyCommand(),
	)
	return cmd
}

// resgister REST routes
func RegisterRoutes(r *mux.Router, indent bool) {
	r.HandleFunc("/keys", QueryKeysRequestHandler(indent)).Methods("GET")
	r.HandleFunc("/keys", AddNewKeyRequestHandler(indent)).Methods("POST")
	r.HandleFunc("/keys/seed", SeedRequestHandler).Methods("GET")
	r.HandleFunc("/keys/{name}/recover", RecoverRequestHandler(indent)).Methods("POST")
	r.HandleFunc("/keys/{name}", GetKeyRequestHandler(indent)).Methods("GET")
	r.HandleFunc("/keys/{name}", UpdateKeyRequestHandler).Methods("PUT")
	r.HandleFunc("/keys/{name}", DeleteKeyRequestHandler).Methods("DELETE")
}
