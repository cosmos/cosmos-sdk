package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin/client/commands"
	rest "github.com/tendermint/basecoin/client/rest"
	coinrest "github.com/tendermint/basecoin/modules/coin/rest"
	rolerest "github.com/tendermint/basecoin/modules/roles/rest"
	"github.com/tendermint/tmlibs/cli"
)

var srvCli = &cobra.Command{
	Use:   "baseserver",
	Short: "Light REST client for tendermint",
	Long:  `Baseserver presents  a nice (not raw hex) interface to the basecoin blockchain structure.`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the light REST client for tendermint",
	Long:  "Access basecoin via REST",
	RunE:  serve,
}

const (
	envPortFlag = "port"
	defaultAlgo = "ed25519"
)

func init() {
	_ = serveCmd.PersistentFlags().Int(envPortFlag, 8998, "the port to run the server on")
}

func serve(cmd *cobra.Command, args []string) error {
	router := mux.NewRouter()

	routeRegistrars := []func(*mux.Router) error{
		// rest.Keys handlers
		rest.NewDefaultKeysManager(defaultAlgo).RegisterAllCRUD,

		// Coin send handler
		coinrest.RegisterCoinSend,
		// Coin query account handler
		coinrest.RegisterQueryAccount,

		// Roles createRole handler
		rolerest.RegisterCreateRole,

		// Basecoin sign transactions handler
		rest.RegisterSignTx,
		// Basecoin post transaction handler
		rest.RegisterPostTx,
	}

	for _, routeRegistrar := range routeRegistrars {
		if err := routeRegistrar(router); err != nil {
			log.Fatal(err)
		}
	}

	port := viper.GetInt(envPortFlag)
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Serving on %q", addr)
	return http.ListenAndServe(addr, router)
}

func main() {
	commands.AddBasicFlags(srvCli)

	srvCli.AddCommand(
		commands.InitCmd,
		commands.VersionCmd,
		serveCmd,
	)

	// this should share the dir with basecli, so you can use the cli and
	// the api interchangeably
	cmd := cli.PrepareMainCmd(srvCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
