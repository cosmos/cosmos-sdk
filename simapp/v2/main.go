package main

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/app/cli"
	"github.com/cosmos/cosmos-sdk/container"
)

func main() {
	cli.Run(
		container.Supply(
			cli.DefaultHome(".simapp"),
		),
		container.Provide(demoCustomAppConfigProvider),
	)
}
