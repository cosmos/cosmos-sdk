package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/cmd/gaiacli/gaiaclicmd"
)

func main() {
	rootCmd := gaiaclicmd.MakeGaiaCLI()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
