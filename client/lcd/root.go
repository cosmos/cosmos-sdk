package lcd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	client "github.com/cosmos/cosmos-sdk/client"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
)

const (
	flagBind = "bind"
	flagCORS = "cors"
)

// ServeCommand will generate a long-running rest server
// (aka Light Client Daemon) that exposes functionality similar
// to the cli, but over rest
func ServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE:  startRESTServer,
	}
	// TODO: handle unix sockets also?
	cmd.Flags().StringP(flagBind, "b", "localhost:1317", "Interface and port that server binds to")
	cmd.Flags().String(flagCORS, "", "Set to domains that can make CORS requests (* for all)")
	cmd.Flags().StringP(client.FlagChainID, "c", "", "ID of chain we connect to")
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	return cmd
}

func startRESTServer(cmd *cobra.Command, args []string) error {
	r := initRouter()

	bind := viper.GetString(flagBind)
	http.ListenAndServe(bind, r)

	return nil
}

func initRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/node_info", rpc.NodeStatusRequestHandler)
	r.HandleFunc("/blocks/{height}", rpc.BlockRequestHandler)
	return r
}
