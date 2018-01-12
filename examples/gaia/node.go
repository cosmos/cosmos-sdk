package main

import (
	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	basecmd "github.com/cosmos/cosmos-sdk/server/commands"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"

	"github.com/cosmos/gaia/modules/stake"
)

// nodeCmd is the entry point for this binary
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "The Cosmos Network delegation-game blockchain test",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func prepareNodeCommands() {

	basecmd.Handler = stack.New(
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
			fee.NewSimpleFeeMiddleware(coin.Coin{"fermion", 0}, fee.Bank),
			stack.Checkpoint{OnDeliver: true},
		).
		Dispatch(
			coin.NewHandler(),
			stack.WrapHandler(roles.NewHandler()),
			stack.WrapHandler(ibc.NewHandler()),
			stake.NewHandler(),
		)

	nodeCmd.AddCommand(
		basecmd.GetInitCmd("fermion", []string{"stake/allowed_bond_denom/fermion"}),
		basecmd.GetTickStartCmd(sdk.TickerFunc(tickFn)),
		basecmd.UnsafeResetAllCmd,
	)
}

// Tick - Called every block even if no transaction, process all queues,
// validator rewards, and calculate the validator set difference
func tickFn(ctx sdk.Context, store state.SimpleDB) (change []*abci.Validator, err error) {
	// first need to prefix the store, at this point it's a global store
	store = stack.PrefixedStore(stake.Name(), store)

	// execute Tick
	change, err = stake.Tick(ctx, store)
	return
}
