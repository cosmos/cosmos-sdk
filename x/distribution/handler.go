package distribution

import (
	metrics "github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgSetWithdrawAddress:
			return handleMsgModifyWithdrawAddress(ctx, msg, k)

		case *types.MsgWithdrawDelegatorReward:
			return handleMsgWithdrawDelegatorReward(ctx, msg, k)

		case *types.MsgWithdrawValidatorCommission:
			return handleMsgWithdrawValidatorCommission(ctx, msg, k)

		case *types.MsgFundCommunityPool:
			return handleMsgFundCommunityPool(ctx, msg, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized distribution message type: %T", msg)
		}
	}
}

// These functions assume everything has been authenticated (ValidateBasic passed, and signatures checked)

func handleMsgModifyWithdrawAddress(ctx sdk.Context, msg *types.MsgSetWithdrawAddress, k keeper.Keeper) (*sdk.Result, error) {
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	withdrawAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	err = k.SetWithdrawAddr(ctx, delegatorAddress, withdrawAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgWithdrawDelegatorReward(ctx sdk.Context, msg *types.MsgWithdrawDelegatorReward, k keeper.Keeper) (*sdk.Result, error) {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	amount, err := k.WithdrawDelegationRewards(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, a := range amount {
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", "withdraw_reward"},
				float32(a.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
			)
		}
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgWithdrawValidatorCommission(ctx sdk.Context, msg *types.MsgWithdrawValidatorCommission, k keeper.Keeper) (*sdk.Result, error) {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	amount, err := k.WithdrawValidatorCommission(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, a := range amount {
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", "withdraw_commission"},
				float32(a.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
			)
		}
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgFundCommunityPool(ctx sdk.Context, msg *types.MsgFundCommunityPool, k keeper.Keeper) (*sdk.Result, error) {
	depositer, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}
	if err := k.FundCommunityPool(ctx, msg.Amount, depositer); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func NewCommunityPoolSpendProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolSpendProposal:
			return keeper.HandleCommunityPoolSpendProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized distr proposal content type: %T", c)
		}
	}
}
