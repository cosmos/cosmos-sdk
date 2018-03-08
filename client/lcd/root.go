package lcd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	wire "github.com/tendermint/go-wire"

	client "github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
	tx "github.com/cosmos/cosmos-sdk/client/tx"
	version "github.com/cosmos/cosmos-sdk/version"
)

const (
	flagBind = "bind"
	flagCORS = "cors"
)

// ServeCommand will generate a long-running rest server
// (aka Light Client Daemon) that exposes functionality similar
// to the cli, but over rest
func ServeCommand(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE:  startRESTServer(cdc),
	}
	// TODO: handle unix sockets also?
	cmd.Flags().StringP(flagBind, "b", "localhost:1317", "Interface and port that server binds to")
	cmd.Flags().String(flagCORS, "", "Set to domains that can make CORS requests (* for all)")
	cmd.Flags().StringP(client.FlagChainID, "c", "", "ID of chain we connect to")
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	return cmd
}

func startRESTServer(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bind := viper.GetString(flagBind)
		r := initRouter(cdc)
		return http.ListenAndServe(bind, r)
	}
}

func initRouter(cdc *wire.Codec) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/version", version.VersionRequestHandler).Methods("GET")
	r.HandleFunc("/node_info", rpc.NodeInfoRequestHandler).Methods("GET")
	r.HandleFunc("/syncing", rpc.NodeSyncingRequestHandler).Methods("GET")
	r.HandleFunc("/keys", keys.QueryKeysRequestHandler).Methods("GET")
	r.HandleFunc("/keys", keys.AddNewKeyRequestHandler).Methods("POST")
	r.HandleFunc("/keys/seed", keys.SeedRequestHandler).Methods("GET")
	r.HandleFunc("/keys/{name}", keys.GetKeyRequestHandler).Methods("GET")
	r.HandleFunc("/keys/{name}", keys.UpdateKeyRequestHandler).Methods("PUT")
	r.HandleFunc("/keys/{name}", keys.DeleteKeyRequestHandler).Methods("DELETE")
	r.HandleFunc("/txs", tx.SearchTxRequestHandler(cdc)).Methods("GET")
	r.HandleFunc("/txs/{hash}", tx.QueryTxRequestHandler(cdc)).Methods("GET")
	r.HandleFunc("/txs/sign", tx.SignTxRequstHandler).Methods("POST")
	r.HandleFunc("/txs/broadcast", tx.BroadcastTxRequestHandler).Methods("POST")
	r.HandleFunc("/blocks/latest", rpc.LatestBlockRequestHandler).Methods("GET")
	r.HandleFunc("/blocks/{height}", rpc.BlockRequestHandler).Methods("GET")
	r.HandleFunc("/validatorsets/latest", rpc.LatestValidatorsetRequestHandler).Methods("GET")
	r.HandleFunc("/validatorsets/{height}", rpc.ValidatorsetRequestHandler).Methods("GET")
	return r
}
