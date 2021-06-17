package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) VerifyInvariant(goCtx context.Context, msg *types.MsgVerifyInvariant) (*types.MsgVerifyInvariantResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	constantFee := sdk.NewCoins(k.GetConstantFee(ctx))

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	if err := k.SendCoinsFromAccountToFeeCollector(ctx, sender, constantFee); err != nil {
		return nil, err
	}

	// use a cached context to avoid gas costs during invariants
	cacheCtx, _ := ctx.CacheContext()

	found := false
	msgFullRoute := msg.FullInvariantRoute()

	var res string
	var stop bool
	for _, invarRoute := range k.Routes() {
		if invarRoute.FullRoute() == msgFullRoute {
			res, stop = invarRoute.Invar(cacheCtx)
			found = true

			break
		}
	}

	if !found {
		return nil, types.ErrUnknownInvariant
	}

	if stop {
		// NOTE currently, because the chain halts here, this transaction will never be included
		// in the blockchain thus the constant fee will have never been deducted. Thus no
		// refund is required.

		// TODO uncomment the following code block with implementation of the circuit breaker
		// // refund constant fee
		// err := k.distrKeeper.DistributeFromFeePool(ctx, constantFee, msg.Sender)
		// if err != nil {
		// // if there are insufficient coins to refund, log the error,
		// // but still halt the chain.
		// logger := ctx.Logger().With("module", "x/crisis")
		// logger.Error(fmt.Sprintf(
		// "WARNING: insufficient funds to allocate to sender from fee pool, err: %s", err))
		//}

		// TODO replace with circuit breaker
		panic(res)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeInvariant,
			sdk.NewAttribute(types.AttributeKeyRoute, msg.InvariantRoute),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCrisis),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgVerifyInvariantResponse{}, nil
}
