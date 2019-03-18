package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
)

// ModuleName is the module name for this module
const ModuleName = "crisis"

func NewHandler(k Keeper, d distr.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {

		switch msg := msg.(type) {
		case MsgVerifyInvariance:
			return handleMsgVerifyInvariance(ctx, msg, k, d)
		default:
			return sdk.ErrTxDecode("invalid message parse in crisis module").Result()
		}
	}
}

func handleMsgVerifyInvariance(ctx sdk.Context, msg MsgVerifyInvariance, k Keeper, d distr.Keeper) sdk.Result {

	// use a cached context to avoid gas costs
	cacheCtx, _ := ctx.CacheContext()

	found := false
	var invarianceErr error
	for _, invarRoute := range k.routes {
		if invarRoute.Route == msg.InvarianceRoute {
			invarianceErr = invarRoute.Invar(cacheCtx)
			found = true
		}
	}
	if !found {
		return ErrUnknownInvariant(DefaultCodespace).Result()
	}

	if invarianceErr != nil {

		// refund constant fee
		refund := k.GetConstantFee(cacheCtx)
		d.DistributeFeePool(ctx, refund, msg.Sender)

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
