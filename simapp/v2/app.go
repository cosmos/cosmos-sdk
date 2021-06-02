package main

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/cli"

	// Load Modules
	_ "github.com/cosmos/cosmos-sdk/x/auth/module"
	_ "github.com/cosmos/cosmos-sdk/x/bank/module"
	_ "github.com/cosmos/cosmos-sdk/x/gov/module"
	_ "github.com/cosmos/cosmos-sdk/x/params/module"
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
