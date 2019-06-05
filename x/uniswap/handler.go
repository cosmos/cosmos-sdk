package uniswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler routes the messages to the handlers
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSwapOrder:
			return HandleMsgSwapOrder(ctx, msg, k)
		case MsgAddLiquidity:
			return HandleMsgAddLiquidity(ctx, msg, k)
		case MsgRemoveLiquidity:
			return HandleMsgRemoveLiquidity(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("unrecognized uniswap message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// HandleMsgSwapOrder handler for MsgSwapOrder
func HandleMsgSwapOrder(ctx sdk.Context, msg types.MsgSwapOrder, k Keeper,
) sdk.Result {

}

// HandleMsgAddLiquidity handler for MsgAddLiquidity
func HandleMsgAddLiquidity(ctx sdk.Context, msg types.MsgAddLiqudity, k Keeper,
) sdk.Result {
	// check if exchange already exists
	totalLiquidity, err := k.GetExchange(ctx, msg.Denom)
	if err != nil {
		k.CreateExchange(ctx, msg.Denom)
	}

}

// HandleMsgRemoveLiquidity handler for MsgRemoveLiquidity
func HandleMsgRemoveLiquidity(ctx sdk.Context, msg types.MsgRemoveLiquidity, k Keeper,
) sdk.Result {

}
