package main

import (
	"os"

	"github.com/pointnetwork/cosmos-point-sdk/server"
	svrcmd "github.com/pointnetwork/cosmos-point-sdk/server/cmd"
	"github.com/pointnetwork/cosmos-point-sdk/simapp"
	"github.com/pointnetwork/cosmos-point-sdk/simapp/simd/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}
