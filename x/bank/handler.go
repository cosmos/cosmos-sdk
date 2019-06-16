package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgSend:
			return handleMsgSend(ctx, k, msg)

		case types.MsgMultiSend:
			return handleMsgMultiSend(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized bank message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, k keeper.Keeper, msg types.MsgSend) sdk.Result {
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}
	err := k.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		types.Category, types.TxCategory,
		types.Sender, msg.FromAddress.String(),
		types.Recipient, msg.ToAddress.String(),
	)

	return sdk.Result{
		Tags: resTags,
	}
}

// Handle MsgMultiSend.
func handleMsgMultiSend(ctx sdk.Context, k keeper.Keeper, msg types.MsgMultiSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}
	resTags, err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	resTags = resTags.AppendTag(types.Category, types.TxCategory)
	return sdk.Result{
		Tags: resTags,
	}
}
