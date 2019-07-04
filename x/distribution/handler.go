package distribution

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSetWithdrawAddress:
			return handleMsgModifyWithdrawAddress(ctx, msg, k)

		case types.MsgWithdrawDelegatorReward:
			return handleMsgWithdrawDelegatorReward(ctx, msg, k)

		case types.MsgWithdrawValidatorCommission:
			return handleMsgWithdrawValidatorCommission(ctx, msg, k)

		default:
			errMsg := fmt.Sprintf("unrecognized distribution message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// These functions assume everything has been authenticated (ValidateBasic passed, and signatures checked)

func handleMsgModifyWithdrawAddress(ctx sdk.Context, msg types.MsgSetWithdrawAddress, k keeper.Keeper) sdk.Result {
	err := k.SetWithdrawAddr(ctx, msg.DelegatorAddress, msg.WithdrawAddress)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgWithdrawDelegatorReward(ctx sdk.Context, msg types.MsgWithdrawDelegatorReward, k keeper.Keeper) sdk.Result {
	_, err := k.WithdrawDelegationRewards(ctx, msg.DelegatorAddress, msg.ValidatorAddress)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgWithdrawValidatorCommission(ctx sdk.Context, msg types.MsgWithdrawValidatorCommission, k keeper.Keeper) sdk.Result {
	_, err := k.WithdrawValidatorCommission(ctx, msg.ValidatorAddress)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func NewCommunityPoolSpendProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Error {
		switch c := content.(type) {
		case types.CommunityPoolSpendProposal:
			return keeper.HandleCommunityPoolSpendProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized distr proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}
