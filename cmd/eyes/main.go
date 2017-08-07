package main

import (
	"os"

	"github.com/tendermint/tmlibs/cli"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/eyes"
	"github.com/tendermint/basecoin/stack"
)

// BuildApp constructs the stack we want to use for this app
func BuildApp() basecoin.Handler {
	return stack.New(
		base.Logger{},
		stack.Recovery{},
	).
		// We do this to demo real usage, also embeds it under it's own namespace
		Dispatch(
			stack.WrapHandler(etc.NewHandler()),
		)
}

func main() {
	rt := commands.RootCmd
	rt.Short = "eyes"
	rt.Long = "A demo app to show key-value store with proofs over abci"

	commands.Handler = BuildApp()

	rt.AddCommand(
		// out own init command to not require argument
		InitCmd,
		commands.StartCmd,
		commands.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(rt, "EYE", os.ExpandEnv("$HOME/.eyes"))
	cmd.Execute()
}
