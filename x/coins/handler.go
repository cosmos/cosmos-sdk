package coins

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handle all "coins" type messages.
// NOTE: Technically, NewHandler only needs a CoinMapper
func NewHandler(cm CoinMapper) sdk.Handler {

	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		cm := CoinMapper{am}
		switch msg := msg.(type) {
		case SendMsg:
			return handleSendMsg(ctx, cm, msg)
		case IssueMsg:
			return handleIssueMsg(ctx, cm, msg)
		default:
			errMsg := "Unrecognized coins Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}

}

// Handle SendMsg.
func handleSendMsg(ctx sdk.Context, cm CoinMapper, msg SendMsg) sdk.Result {

	fromAccount := ctx, cm.am.GetAccount(ctx, msg.FromAddress)
	if fromAccount == nil {
		return ErrUnknownAddress(msg.FromAddress.String())
	}
	
	fromAccount := cm.GetCoinAccount(fromAccount)
	if fromAccount == nil {
		return ErrUnknownAddress(msg.FromAddress.String())
	}


	toAccount := cm.am.GetAccount(ctx, msg.ToAccount)
	if toAccount == nil {
		toAccount, err := cm.am.makeAccount(msg.ToAddress)
		if err != nil {
			return err.Result()
		}
		am.SetAccount(ctx, )
	} else {

	}

	for _, in := range msg.Inputs {
		_, err := cm.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err.Result()
		}
	}

	for _, out := range msg.Outputs {
		_, err := cm.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{} // TODO
}

// Handle IssueMsg.
func handleIssueMsg(ctx sdk.Context, cm CoinMapper, msg IssueMsg) sdk.Result {
	panic("not implemented yet")
}
