package main

import (
	"fmt"
	"os"

	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"
	servercore "cosmossdk.io/core/server"
)

func main() {
	rootCmd := cmd.NewRootCmd[servercore.AppI[transaction.Tx], transaction.Tx]()
	if err := serverv2.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
