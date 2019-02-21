package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "crisis"

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {

		switch msg := msg.(type) {
		case MsgVerifyInvariance:
			return handleMsgVerifyInvariance(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in crisis module").Result()
		}
	}
}

func handleMsgVerifyInvariance(ctx sdk.Context, msg MsgVerifyInvariance, k Keeper) sdk.Result {

	// get the initial gas consumption level, for refund comparison
	initGas := ctx.GasMeter().GasConsumed()

	found := false
	for _, invarRoutes := range k.routes {
		if invarRoutes.Route == msg.InvarianceRoute {
			invarianceErr := invarRoutes.Invariant()
			found = true
		}
	}
	if !found {
		return ErrUnknownInvariant(DefaultCodespace).Result()
	}

	if invarianceErr != nil {

		// refund gas
		finalGas := ctx.GasMeter().GasConsumed()
		diffGas := finalGas - initGas
		gasPrice := ctx.TxGasPrice()
		refund, _ := gasPrice.MulDec(NewDec(int64(diffGas))).TruncateDecimal() // TODO verify if should round up
		distrKeeper.DistributeFeePool(ctx, refund, msg.Sender)

		// TODO replace with circuit breaker
		panic(invarianceErr)
	}

	tags := sdk.NewTags(
		"invariant", msg.InvarianceRoute,
	)
	return sdk.Result{
		Tags: tags,
	}
}
