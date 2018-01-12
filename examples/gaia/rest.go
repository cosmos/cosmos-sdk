package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands"
	rest "github.com/cosmos/cosmos-sdk/client/rest"
	coinrest "github.com/cosmos/cosmos-sdk/modules/coin/rest"
	noncerest "github.com/cosmos/cosmos-sdk/modules/nonce/rest"
	rolerest "github.com/cosmos/cosmos-sdk/modules/roles/rest"

	stakerest "github.com/cosmos/gaia/modules/stake/rest"
)

const defaultAlgo = "ed25519"

var (
	restServerCmd = &cobra.Command{
		Use:   "rest-server",
		Short: "REST client for gaia commands",
		Long:  `Gaiaserver presents  a nice (not raw hex) interface to the gaia blockchain structure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdRestServer(cmd, args)
		},
	}

	flagPort = "port"
)

func prepareRestServerCommands() {
	commands.AddBasicFlags(restServerCmd)
	restServerCmd.PersistentFlags().IntP(flagPort, "p", 8998, "port to run the server on")
}

func cmdRestServer(cmd *cobra.Command, args []string) error {
	router := mux.NewRouter()

	rootDir := viper.GetString(cli.HomeFlag)
	keyMan := client.GetKeyManager(rootDir)
	serviceKeys := rest.NewServiceKeys(keyMan)
	serviceTxs := rest.NewServiceTxs(commands.GetNode())

	routeRegistrars := []func(*mux.Router) error{
		// rest.Keys handlers
		serviceKeys.RegisterCRUD,

		// Coin handlers (Send, Query, SearchSent)
		coinrest.RegisterAll,

		// Roles createRole handler
		rolerest.RegisterCreateRole,

		// Gaia sign transactions handler
		serviceKeys.RegisterSignTx,
		// Gaia post transaction handler
		serviceTxs.RegisterPostTx,

		// Nonce query handler
		noncerest.RegisterQueryNonce,

		// Staking query handlers
		stakerest.RegisterQueryCandidate,
		stakerest.RegisterQueryCandidates,
		stakerest.RegisterQueryDelegatorBond,
		stakerest.RegisterQueryDelegatorCandidates,
		// Staking tx builders
		stakerest.RegisterDelegate,
		stakerest.RegisterUnbond,
	}

	for _, routeRegistrar := range routeRegistrars {
		if err := routeRegistrar(router); err != nil {
			log.Fatal(err)
		}
	}

	addr := fmt.Sprintf(":%d", viper.GetInt(flagPort))

	log.Printf("Serving on %q", addr)
	return http.ListenAndServe(addr, router)
}
