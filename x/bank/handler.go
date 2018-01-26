package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

// Handle all "bank" type messages.
func NewHandler(am sdk.AccountMapper) sdk.Handler {

	return func(ctx sdk.Context, tx sdk.Tx) sdk.Result {
		cm := CoinMapper{am}
		msg := tx.(sdk.Msg)
		switch msg := msg.(type) {
		case SendMsg:
			return handleSendMsg(ctx, cm, msg)
		case IssueMsg:
			return handleIssueMsg(ctx, cm, msg)
		default:
			errMsg := "Unrecognized bank Tx type: " + reflect.TypeOf(tx).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}

}

// Handle SendMsg.
func handleSendMsg(ctx sdk.Context, cm CoinMapper, msg SendMsg) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked

	for _, in := range msg.Inputs {
		_, err := cm.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return ErrInvalidInput("").TraceCause(err, "").Result()
		}
	}

	for _, out := range msg.Outputs {
		_, err := cm.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return ErrInvalidOutput("").TraceCause(err, "").Result()
		}
	}

	return sdk.Result{} // TODO
}

// Handle IssueMsg.
func handleIssueMsg(ctx sdk.Context, cm CoinMapper, msg IssueMsg) sdk.Result {
	panic("not implemented yet")
}
