package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/simapp/simd/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
