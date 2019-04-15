package crisis

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/tags"
)

// RouterKey
const RouterKey = ModuleName

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgVerifyInvariant:
			return handleMsgVerifyInvariant(ctx, msg, k)

		default:
			errMsg := fmt.Sprintf("unrecognized crisis message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgVerifyInvariant(ctx sdk.Context, msg MsgVerifyInvariant, k Keeper) sdk.Result {

	// remove the constant fee
	constantFee := sdk.NewCoins(k.GetConstantFee(ctx))
	_, err := k.bankKeeper.SubtractCoins(ctx, msg.Sender, constantFee)
	if err != nil {
		return err.Result()
	}
	_ = k.feeCollectionKeeper.AddCollectedFees(ctx, constantFee)

	// use a cached context to avoid gas costs during invariants
	cacheCtx, _ := ctx.CacheContext()

	found := false
	var invarianceErr error
	msgFullRoute := msg.FullInvariantRoute()
	for _, invarRoute := range k.routes {
		if invarRoute.FullRoute() == msgFullRoute {
			invarianceErr = invarRoute.Invar(cacheCtx)
			found = true
			break
		}
	}
	if !found {
		return ErrUnknownInvariant(DefaultCodespace).Result()
	}

	if invarianceErr != nil {

		// NOTE currently, because the chain halts here, this transaction will never be included
		// in the blockchain thus the constant fee will have never been deducted. Thus no
		// refund is required.

		// TODO uncomment the following code block with implementation of the circuit breaker
		//// refund constant fee
		//err := k.distrKeeper.DistributeFeePool(ctx, constantFee, msg.Sender)
		//if err != nil {
		//// if there are insufficient coins to refund, log the error,
		//// but still halt the chain.
		//logger := ctx.Logger().With("module", "x/crisis")
		//logger.Error(fmt.Sprintf(
		//"WARNING: insufficient funds to allocate to sender from fee pool, err: %s", err))
		//}

		// TODO replace with circuit breaker
		panic(invarianceErr)
	}

	resTags := sdk.NewTags(
		tags.Sender, msg.Sender.String(),
		tags.Invariant, msg.InvariantRoute,
	)
	return sdk.Result{
		Tags: resTags,
	}
}
