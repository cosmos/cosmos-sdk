package main

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/app/cli"

	"github.com/cosmos/cosmos-sdk/app"
)

//go:embed app.yaml
var configYaml []byte

func main() {
	config, err := app.ReadYAMLConfig(configYaml)
	if err != nil {
		panic(err)
	}

	cli.Run(cli.Options{
		DefaultAppConfig: config,
		DefaultHome:      "simapp",
		EnvPrefix:        "SIMAPP",
	})
}
