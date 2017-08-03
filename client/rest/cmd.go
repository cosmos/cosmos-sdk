package rest

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	coinrest "github.com/tendermint/basecoin/modules/coin/rest"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the light REST client for tendermint",
	Long:  "Access basecoin via REST",
	RunE:  serve,
}

const envPortFlag = "port"

func init() {
	_ = ServeCmd.PersistentFlags().Int(envPortFlag, 8998, "the port to run the server on")
}

const defaultAlgo = "ed25519"

func serve(cmd *cobra.Command, args []string) error {
	port := viper.GetInt(envPortFlag)
	keysManager := DefaultKeysManager()
	router := mux.NewRouter()
	ctx := Context{
		Keys: New(keysManager, defaultAlgo),
	}
	if err := ctx.RegisterHandlers(router); err != nil {
		return err
	}
	if err := coinrest.RegisterHandlers(router); err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Serving on %q", addr)
	return http.ListenAndServe(addr, router)
}
