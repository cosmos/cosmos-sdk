package staking

import (
	"time"

	"github.com/armon/go-metrics"
	gogotypes "github.com/gogo/protobuf/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/KiraCore/cosmos-sdk/telemetry"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	sdkerrors "github.com/KiraCore/cosmos-sdk/types/errors"
	"github.com/KiraCore/cosmos-sdk/x/staking/keeper"
	"github.com/KiraCore/cosmos-sdk/x/staking/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgCreateValidator:
			return handleMsgCreateValidator(ctx, msg, k)

		case *types.MsgEditValidator:
			return handleMsgEditValidator(ctx, msg, k)

		case *types.MsgDelegate:
			return handleMsgDelegate(ctx, msg, k)

		case *types.MsgBeginRedelegate:
			return handleMsgBeginRedelegate(ctx, msg, k)

		case *types.MsgUndelegate:
			return handleMsgUndelegate(ctx, msg, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgCreateValidator(ctx sdk.Context, msg *types.MsgCreateValidator, k keeper.Keeper) (*sdk.Result, error) {
	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, msg.ValidatorAddress); found {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, msg.Pubkey)
	if err != nil {
		return nil, err
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); found {
		return nil, types.ErrValidatorPubKeyExists
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Value.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(types.ErrBadDenom, "got %s, expected %s", msg.Value.Denom, bondDenom)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	cp := ctx.ConsensusParams()
	if cp != nil && cp.Validator != nil {
		tmPubKey := tmtypes.TM2PB.PubKey(pk)

		if !tmstrings.StringInSlice(tmPubKey.Type, cp.Validator.PubKeyTypes) {
			return nil, sdkerrors.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", tmPubKey.Type, cp.Validator.PubKeyTypes,
			)
		}
	}

	validator := types.NewValidator(msg.ValidatorAddress, pk, msg.Description)
	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	validator.MinSelfDelegation = msg.MinSelfDelegation

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.OperatorAddress)

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Delegate(ctx, msg.DelegatorAddress, msg.Value.Amount, sdk.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgEditValidator(ctx sdk.Context, msg *types.MsgEditValidator, k keeper.Keeper) (*sdk.Result, error) {
	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	validator.Description = description

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return nil, err
		}

		// call the before-modification hook since we're about to update the commission
		k.BeforeValidatorModified(ctx, msg.ValidatorAddress)

		validator.Commission = commission
	}

	if msg.MinSelfDelegation != nil {
		if !msg.MinSelfDelegation.GT(validator.MinSelfDelegation) {
			return nil, types.ErrMinSelfDelegationDecreased
		}

		if msg.MinSelfDelegation.GT(validator.Tokens) {
			return nil, types.ErrSelfDelegationBelowMinimum
		}

		validator.MinSelfDelegation = (*msg.MinSelfDelegation)
	}

	k.SetValidator(ctx, validator)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
			sdk.NewAttribute(types.AttributeKeyMinSelfDelegation, validator.MinSelfDelegation.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgDelegate(ctx sdk.Context, msg *types.MsgDelegate, k keeper.Keeper) (*sdk.Result, error) {
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(types.ErrBadDenom, "got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	// NOTE: source funds are always unbonded
	_, err := k.Delegate(ctx, msg.DelegatorAddress, msg.Amount.Amount, sdk.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounter(1, types.ModuleName, "delegate")
		telemetry.SetGaugeWithLabels(
			[]string{"tx", "msg", msg.Type()},
			float32(msg.Amount.Amount.Int64()),
			[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
		)
	}()

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgUndelegate(ctx sdk.Context, msg *types.MsgUndelegate, k keeper.Keeper) (*sdk.Result, error) {
	shares, err := k.ValidateUnbondAmount(
		ctx, msg.DelegatorAddress, msg.ValidatorAddress, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(types.ErrBadDenom, "got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	completionTime, err := k.Undelegate(ctx, msg.DelegatorAddress, msg.ValidatorAddress, shares)
	if err != nil {
		return nil, err
	}

	ts, err := gogotypes.TimestampProto(completionTime)
	if err != nil {
		return nil, types.ErrBadRedelegationAddr
	}

	defer func() {
		telemetry.IncrCounter(1, types.ModuleName, "undelegate")
		telemetry.SetGaugeWithLabels(
			[]string{"tx", "msg", msg.Type()},
			float32(msg.Amount.Amount.Int64()),
			[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
		)
	}()

	completionTimeBz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(ts)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Data: completionTimeBz, Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgBeginRedelegate(ctx sdk.Context, msg *types.MsgBeginRedelegate, k keeper.Keeper) (*sdk.Result, error) {
	shares, err := k.ValidateUnbondAmount(
		ctx, msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(types.ErrBadDenom, "got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	completionTime, err := k.BeginRedelegation(
		ctx, msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.ValidatorDstAddress, shares,
	)
	if err != nil {
		return nil, err
	}

	ts, err := gogotypes.TimestampProto(completionTime)
	if err != nil {
		return nil, types.ErrBadRedelegationAddr
	}

	defer func() {
		telemetry.IncrCounter(1, types.ModuleName, "redelegate")
		telemetry.SetGaugeWithLabels(
			[]string{"tx", "msg", msg.Type()},
			float32(msg.Amount.Amount.Int64()),
			[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
		)
	}()

	completionTimeBz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(ts)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress.String()),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Data: completionTimeBz, Events: ctx.EventManager().ABCIEvents()}, nil
}
