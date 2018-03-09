package lcd

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

const (
	flagBind             = "bind"
	flagCORS             = "cors"
	flagUnsafeConnection = "unsafe_connection"
)

// ServeCommand will generate a long-running rest server
// (aka Light Client Daemon) that exposes functionality similar
// to the cli, but over rest
func ServeCommand() *cobra.Command {
	// TODO get code from app
	cdc := wire.NewCodec()
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE:  startRESTServer(cdc),
	}
	// TODO: handle unix sockets also?
	cmd.Flags().StringP(flagBind, "b", "localhost:1317", "Interface and port that server binds to")
	cmd.Flags().String(flagCORS, "", "Set to domains that can make CORS requests (* for all)")
	cmd.Flags().String(flagUnsafeConnection, "false", "Do not enable HTTPS for the REST server")
	cmd.Flags().StringP(client.FlagChainID, "c", "", "ID of chain we connect to")
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	return cmd
}

func startRESTServer(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bind := viper.GetString(flagBind)
		unsafeConnection, err := strconv.ParseBool(viper.GetString(flagUnsafeConnection))
		if err != nil {
			return err
		}
		r := initRouter(cdc)

		if unsafeConnection {
			return http.ListenAndServe(bind, r)
		}

		// setup https
		// get path to certificates
		ex, err := os.Executable()
		if err != nil {
			return err
		}
		exPath := filepath.Dir(ex)

		return http.ListenAndServeTLS(bind, filepath.Join(exPath, "server.crt"), filepath.Join(exPath, "server.key"), r)
	}
}

func initRouter(cdc *wire.Codec) http.Handler {
	r := mux.NewRouter()

	// register routes here
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("alive"))
	})

	return r
}
