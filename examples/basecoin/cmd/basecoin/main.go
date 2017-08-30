package main

import (
	"os"

	"github.com/tendermint/tmlibs/cli"

	sdk "github.com/cosmos/cosmos-sdk"
	client "github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/cmd/basecoin/commands"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"
)

// BuildApp constructs the stack we want to use for this app
func BuildApp(feeDenom string) sdk.Handler {
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
			coin.NewHandler(),
			stack.WrapHandler(roles.NewHandler()),
			stack.WrapHandler(ibc.NewHandler()),
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
		client.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(rt, "BC", os.ExpandEnv("$HOME/.basecoin"))
	cmd.Execute()
}
