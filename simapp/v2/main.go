package main

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/container"

	"github.com/cosmos/cosmos-sdk/app/cli"

	// Load Modules
	_ "github.com/cosmos/cosmos-sdk/x/auth/module"
)

//go:embed app.yaml
var configYaml []byte

func main() {
	cli.Run(
		container.Supply(
			app.Name("simapp"),
			cli.DefaultHome(".simapp"),
		),
		app.ProvideAppConfigYAML(configYaml),
	)
}
