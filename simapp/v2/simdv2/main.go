package main

import (
	"fmt"
	"os"

	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := serverv2.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
