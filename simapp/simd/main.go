package main

import (
	"os"

	"github.com/KiraCore/cosmos-sdk/simapp/simd/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
