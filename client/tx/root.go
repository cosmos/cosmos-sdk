package tx

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

// AddCommands adds a number of tx-query related subcommands
func AddCommands(cmd *cobra.Command, cdc *wire.Codec) {
	cmd.AddCommand(
		SearchTxCmd(cdc),
		QueryTxCmd(cdc),
	)
}

// register REST routes
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc("/txs/{hash}", QueryTxRequestHandlerFn(cdc, ctx)).Methods("GET")
	r.HandleFunc("/txs", SearchTxRequestHandlerFn(ctx, cdc)).Methods("GET")
	// r.HandleFunc("/txs/sign", SignTxRequstHandler).Methods("POST")
	// r.HandleFunc("/txs/broadcast", BroadcastTxRequestHandler).Methods("POST")
}
