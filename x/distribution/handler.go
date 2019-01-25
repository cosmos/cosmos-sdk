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
		case types.MsgSetWithdrawAddress:
			return handleMsgModifyWithdrawAddress(ctx, msg, k)
		case types.MsgWithdrawDelegatorReward:
			return handleMsgWithdrawDelegatorReward(ctx, msg, k)
		case types.MsgWithdrawValidatorCommission:
			return handleMsgWithdrawValidatorCommission(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in distribution module").Result()
		}
	}
}

// These functions assume everything has been authenticated (ValidateBasic passed, and signatures checked)

func handleMsgModifyWithdrawAddress(ctx sdk.Context, msg types.MsgSetWithdrawAddress, k keeper.Keeper) sdk.Result {

	err := k.SetWithdrawAddr(ctx, msg.DelegatorAddr, msg.WithdrawAddr)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgWithdrawDelegatorReward(ctx sdk.Context, msg types.MsgWithdrawDelegatorReward, k keeper.Keeper) sdk.Result {

	err := k.WithdrawDelegationRewards(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.Validator, []byte(msg.ValidatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgWithdrawValidatorCommission(ctx sdk.Context, msg types.MsgWithdrawValidatorCommission, k keeper.Keeper) sdk.Result {

	err := k.WithdrawValidatorCommission(ctx, msg.ValidatorAddr)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Validator, []byte(msg.ValidatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}
