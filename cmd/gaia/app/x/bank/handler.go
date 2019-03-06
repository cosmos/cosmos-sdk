// Package bank contains a forked version of the bank module. It only contains
// a modified message handler to support a very limited form of transfers during
// mainnet launch -- MsgMultiSend messages.
//
// NOTE: This fork should be removed entirely once transfers are enabled and
// the Gaia router should be reset to using the original bank module handler.
package bank

import (
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

var (
	uatomDenom    = "uatom"
	atomsToUatoms = int64(1000000)

	// BurnedCoinsAccAddr represents the burn account address used for
	// MsgMultiSend message during the period for which transfers are disabled.
	// Its Bech32 address is cosmos1x4p90uuy63fqzsheamn48vq88q3eusykf0a69v.
	BurnedCoinsAccAddr = sdk.AccAddress(crypto.AddressHash([]byte("bankBurnedCoins")))
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k bank.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case bank.MsgSend:
			return handleMsgSend(ctx, k, msg)

		case bank.MsgMultiSend:
			return handleMsgMultiSend(ctx, k, msg)

		default:
			errMsg := "Unrecognized bank Msg type: %s" + msg.Type()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// handleMsgSend implements a MsgSend message handler. It operates no differently
// than the standard bank module MsgSend message handler in that it transfers
// an amount from one account to another under the condition of transfers being
// enabled.
func handleMsgSend(ctx sdk.Context, k bank.Keeper, msg bank.MsgSend) sdk.Result {
	// No need to modify handleMsgSend as the forked module requires no changes,
	// so we can just call the standard bank modules handleMsgSend since we know
	// the message is of type MsgSend.
	return bank.NewHandler(k)(ctx, msg)
}

// handleMsgMultiSend implements a modified forked version of a MsgMultiSend
// message handler. If transfers are disabled, a modified version of MsgMultiSend
// is allowed where there must be a single input and only two outputs. The first
// of the two outputs must be to a specific burn address defined by
// burnedCoinsAccAddr. In addition, the output amounts must be of 9atom and
// 1uatom respectively.
func handleMsgMultiSend(ctx sdk.Context, k bank.Keeper, msg bank.MsgMultiSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked
	if !k.GetSendEnabled(ctx) {
		if !validateMultiSendTransfersDisabled(msg) {
			return bank.ErrSendDisabled(k.Codespace()).Result()
		}
	}

	tags, err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: tags,
	}
}

func validateMultiSendTransfersDisabled(msg bank.MsgMultiSend) bool {
	nineAtoms := sdk.Coins{sdk.NewInt64Coin(uatomDenom, 9*atomsToUatoms)}
	oneAtom := sdk.Coins{sdk.NewInt64Coin(uatomDenom, 1*atomsToUatoms)}

	if len(msg.Inputs) != 1 {
		return false
	}
	if len(msg.Outputs) != 2 {
		return false
	}

	if !msg.Outputs[0].Address.Equals(BurnedCoinsAccAddr) {
		return false
	}
	if !msg.Outputs[0].Coins.IsEqual(nineAtoms) {
		return false
	}
	if !msg.Outputs[1].Coins.IsEqual(oneAtom) {
		return false
	}

	return true
}
