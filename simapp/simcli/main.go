package main

import (
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/simapp"
)

func main() {
	err := autocli.RunFromAppConfig(simapp.AppConfig)
	if err != nil {
		panic(err)
	}
}
