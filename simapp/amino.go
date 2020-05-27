// +build test_amino

package simapp

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func NewAnteHandler(ak auth.AccountKeeper, bk bank.Keeper, ibcK ibc.Keeper) types.AnteHandler {
	return ante.NewAnteHandler(
		ak, bk, ibcK, ante.DefaultSigVerificationGasConsumer,
	)
}
