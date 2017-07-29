package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/client/commands"
	rest "github.com/tendermint/basecoin/client/rest"
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

var port int

func main() {
	commands.AddBasicFlags(srvCli)

	flagset := serveCmd.Flags()
	flagset.IntVar(&port, "port", 8998, "the port to run the server on")

	srvCli.AddCommand(
		commands.InitCmd,
		serveCmd,
	)

	// TODO: Decide whether to use $HOME/.basecli for compatibility
	// or just use $HOME/.baseserver?
	cmd := cli.PrepareMainCmd(srvCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

const defaultAlgo = "ed25519"

func serve(cmd *cobra.Command, args []string) error {
	keysManager := rest.DefaultKeysManager()
	router := mux.NewRouter()
	ctx := rest.Context{
		Keys: rest.New(keysManager, defaultAlgo),
	}
	if err := ctx.RegisterHandlers(router); err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Serving on %q", addr)
	return http.ListenAndServe(addr, router)
}
