package main

import (
	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/simd/cmd"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		panic(err)
	}
}
