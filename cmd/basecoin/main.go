package main

import (
	"os"

	"github.com/tendermint/tmlibs/cli"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/ibc"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
)

// BuildApp constructs the stack we want to use for this app
func BuildApp(feeDenom string) basecoin.Handler {
	// use the default stack
	c := coin.NewHandler()
	r := roles.NewHandler()
	i := ibc.NewHandler()

	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
		stack.Checkpoint{OnCheck: true},
		nonce.ReplayCheck{},
	).
		IBC(ibc.NewMiddleware()).
		Apps(
			roles.NewMiddleware(),
			fee.NewSimpleFeeMiddleware(coin.Coin{feeDenom, 0}, fee.Bank),
			stack.Checkpoint{OnDeliver: true},
		).
		Dispatch(
			stack.WrapHandler(c),
			stack.WrapHandler(r),
			stack.WrapHandler(i),
		)
}

func main() {
	rt := commands.RootCmd

	// require all fees in mycoin - change this in your app!
	commands.Handler = BuildApp("mycoin")

	rt.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		//commands.RelayCmd,
		commands.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(rt, "BC", os.ExpandEnv("$HOME/.basecoin"))
	cmd.Execute()
}
