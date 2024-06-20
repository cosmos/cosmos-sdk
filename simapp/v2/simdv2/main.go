package main

import (
	"fmt"
	"os"

	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd" // TODO(@julienrbrt), no need to abstract this.
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
