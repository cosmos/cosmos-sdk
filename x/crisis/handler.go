package crisis

import (
	"fmt"

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
	fmt.Println("wackydebugoutput handleMsgVerifyInvariant 0")

	// remove the constant fee
	constantFee := sdk.NewCoins(k.GetConstantFee(ctx))
	_, _, err := k.bankKeeper.SubtractCoins(ctx, msg.Sender, constantFee)
	if err != nil {
		fmt.Println("wackydebugoutput handleMsgVerifyInvariant 1")
		return err.Result()
	}
	fmt.Println("wackydebugoutput handleMsgVerifyInvariant 2")
	_ = k.feeCollectionKeeper.AddCollectedFees(ctx, constantFee)

	// use a cached context to avoid gas costs during invariants
	cacheCtx, _ := ctx.CacheContext()

	found := false
	var invarianceErr error
	for _, invarRoute := range k.routes {
		fmt.Println("wackydebugoutput handleMsgVerifyInvariant 3")
		if invarRoute.Route == msg.InvariantRoute {
			fmt.Println("wackydebugoutput handleMsgVerifyInvariant 4")
			invarianceErr = invarRoute.Invar(cacheCtx)
			found = true
		}
		fmt.Println("wackydebugoutput handleMsgVerifyInvariant 5")
	}
	fmt.Println("wackydebugoutput handleMsgVerifyInvariant 6")
	if !found {
		fmt.Println("wackydebugoutput handleMsgVerifyInvariant 7")
		return ErrUnknownInvariant(DefaultCodespace).Result()
	}
	fmt.Println("wackydebugoutput handleMsgVerifyInvariant 8")

	if invarianceErr != nil {
		fmt.Println("wackydebugoutput handleMsgVerifyInvariant 9")

		// refund constant fee
		err := k.distrKeeper.DistributeFeePool(ctx, constantFee, msg.Sender)
		if err != nil {
			fmt.Printf("debug err: %v\n", err)
			fmt.Println("wackydebugoutput handleMsgVerifyInvariant 10")
			return err.Result()
		}
		fmt.Println("wackydebugoutput handleMsgVerifyInvariant 11")

		// TODO replace with circuit breaker
		panic(invarianceErr)
	}
	fmt.Println("wackydebugoutput handleMsgVerifyInvariant 12")

	tags := sdk.NewTags(
		"invariant", msg.InvariantRoute,
	)
	return sdk.Result{
		Tags: tags,
	}
}
