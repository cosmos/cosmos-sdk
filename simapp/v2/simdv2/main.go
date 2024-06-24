package main

import (
	"fmt"
	"os"

	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd" // TODO(@julienrbrt), no need to abstract this.
)

func main() {
	rootCmd := cmd.NewRootCmd[serverv2.AppI[transaction.Tx], transaction.Tx]()
	if err := svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
