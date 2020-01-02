package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/crisis/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/crisis/internal/types"
)

// RouterKey
const RouterKey = types.ModuleName

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgVerifyInvariant:
			return handleMsgVerifyInvariant(ctx, msg, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized crisis message type: %T", msg)
		}
	}
}

func handleMsgVerifyInvariant(ctx sdk.Context, msg types.MsgVerifyInvariant, k keeper.Keeper) (*sdk.Result, error) {
	constantFee := sdk.NewCoins(k.GetConstantFee(ctx))

	if err := k.SendCoinsFromAccountToFeeCollector(ctx, msg.Sender, constantFee); err != nil {
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
		//// refund constant fee
		//err := k.distrKeeper.DistributeFromFeePool(ctx, constantFee, msg.Sender)
		//if err != nil {
		//// if there are insufficient coins to refund, log the error,
		//// but still halt the chain.
		//logger := ctx.Logger().With("module", "x/crisis")
		//logger.Error(fmt.Sprintf(
		//"WARNING: insufficient funds to allocate to sender from fee pool, err: %s", err))
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
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
