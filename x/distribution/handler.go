package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/tags"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case types.MsgModifyWithdrawAddress:
			return handleMsgModifyWithdrawAddress(ctx, msg, k)
		case types.MsgWithdrawDelegatorRewardsAll:
			return handleMsgWithdrawDelegatorRewardsAll(ctx, msg, k)
		case types.MsgWithdrawDelegatorReward:
			return handleMsgWithdrawDelegatorReward(ctx, msg, k)
		case types.MsgWithdrawValidatorRewardsAll:
			return handleMsgWithdrawValidatorRewardsAll(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in distribution module").Result()
		}
	}
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgModifyWithdrawAddress(ctx sdk.Context, msg types.MsgModifyWithdrawAddress, k keeper.Keeper) sdk.Result {

	k.SetDelegatorWithdrawAddr(ctx, msg.DelegatorAddr, msg.WithdrawAddress)

	tags := sdk.NewTags(
		tags.Action, tags.ActionModifyWithdrawAddress,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgWithdrawDelegatorRewardsAll(ctx sdk.Context, msg types.MsgWithdrawDelegatorRewardsAll, k keeper.Keeper) sdk.Result {

	k.WithdrawDelegationRewardsAll(ctx, msg.DelegatorAddr)

	tags := sdk.NewTags(
		tags.Action, tags.ActionWithdrawDelegatorRewardsAll,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgWithdrawDelegatorReward(ctx sdk.Context, msg types.MsgWithdrawDelegatorReward, k keeper.Keeper) sdk.Result {

	k.WithdrawDelegationReward(ctx, msg.DelegatorAddr, msg.ValidatorAddr)

	tags := sdk.NewTags(
		tags.Action, tags.ActionWithdrawDelegatorReward,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.Validator, []byte(msg.ValidatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgWithdrawValidatorRewardsAll(ctx sdk.Context, msg types.MsgWithdrawValidatorRewardsAll, k keeper.Keeper) sdk.Result {

	k.WithdrawValidatorRewardsAll(ctx, msg.ValidatorAddr)

	tags := sdk.NewTags(
		tags.Action, tags.ActionWithdrawValidatorRewardsAll,
		tags.Validator, []byte(msg.ValidatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}
