package main

import (
	"os"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/simd/cmd"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		log.NewLogger(rootCmd.OutOrStderr()).Error("failure when running app", "err", err)
		os.Exit(1)
	}
}
