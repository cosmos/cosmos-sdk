package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ModuleName is the module name for this module
const (
	ModuleName = "crisis"
	RouterKey  = ModuleName
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {

		switch msg := msg.(type) {
		case MsgVerifyInvariant:
			return handleMsgVerifyInvariant(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in crisis module").Result()
		}
	}
}

func handleMsgVerifyInvariant(ctx sdk.Context, msg MsgVerifyInvariant, k Keeper) sdk.Result {

	// remove the constant fee
	constantFee := sdk.NewCoins(k.GetConstantFee(ctx))
	_, _, err := k.bankKeeper.SubtractCoins(ctx, msg.Sender, constantFee)
	if err != nil {
		return err.Result()
	}
	_ = k.feeCollectionKeeper.AddCollectedFees(ctx, constantFee)

	// use a cached context to avoid gas costs during invariants
	cacheCtx, _ := ctx.CacheContext()

	found := false
	var invarianceErr error
	for _, invarRoute := range k.routes {
		if invarRoute.Route == msg.InvariantRoute {
			invarianceErr = invarRoute.Invar(cacheCtx)
			found = true
		}
	}
	if !found {
		return ErrUnknownInvariant(DefaultCodespace).Result()
	}

	if invarianceErr != nil {

		// refund constant fee
		k.distrKeeper.DistributeFeePool(ctx, constantFee, msg.Sender)

		// TODO replace with circuit breaker
		panic(invarianceErr)
	}

	tags := sdk.NewTags(
		"invariant", msg.InvariantRoute,
	)
	return sdk.Result{
		Tags: tags,
	}
}
