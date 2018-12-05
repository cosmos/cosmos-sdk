package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/cmd/gaiad/gaiadcmd"
)

func main() {
	rootCmd := gaiadcmd.MakeGaiad()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
