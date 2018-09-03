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
		addKeyCommand(),
		listKeysCmd,
		showKeysCmd(),
		signCommand(),
		client.LineBreak,
		deleteKeyCommand(),
		updateKeyCommand(),
	)
	return cmd
}

// resgister REST routes
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/keys", QueryKeysRequestHandler).Methods("GET")
	r.HandleFunc("/keys", AddNewKeyRequestHandler).Methods("POST")
	r.HandleFunc("/keys/{name}/recover", RecoverKeyResuestHandler).Methods("POST")
	r.HandleFunc("/keys/{name}/sign", SignResuest).Methods("POST")
	r.HandleFunc("/keys/{name}", GetKeyRequestHandler).Methods("GET")
	r.HandleFunc("/keys/{name}", UpdateKeyRequestHandler).Methods("PUT")
	r.HandleFunc("/keys/{name}", DeleteKeyRequestHandler).Methods("DELETE")
}
