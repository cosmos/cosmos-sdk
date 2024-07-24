package main

import (
	"fmt"
	"os"

	clientv2helpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd[transaction.Tx]()
	if err := serverv2.Execute(rootCmd, clientv2helpers.EnvPrefix, simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
