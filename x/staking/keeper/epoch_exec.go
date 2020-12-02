package keeper

import (
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// EpochCreateValidatorSelfDelegation does do self-delegation
func (k Keeper) EpochCreateValidatorSelfDelegation(ctx sdk.Context, msg *types.MsgCreateValidator) error {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	// check to see if the pubkey or sender has been registered before
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded

	// Thoughts: since delegation goes to Unbonded pool, as long as we only run validator set update on epochs, it won't affect
	// anything.
	// Warn: Within slashing module, we should update only partial validator set instead of full since it can automatically
	// set newly created validators to Bonded status
	// I think keeping this as it is is quite good as delegators are not needed to wait for validator to be created
	// on epochs but just delegate to validators that is going to be activated on next epoch
	_, err = k.Delegate(ctx, delegatorAddress, msg.Value.Amount, types.Unbonded, validator, true)
	if err != nil {
		return err
	}

	return nil
}

// EpochEditValidator logic is moved from msgServer.EditValidator
func (k Keeper) EpochEditValidator(ctx sdk.Context, msg *types.MsgEditValidator) error {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}
	// validator must already be registered
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return err
	}

	validator.Description = description

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return err
		}

		// call the before-modification hook since we're about to update the commission
		k.BeforeValidatorModified(ctx, valAddr)

		validator.Commission = commission
	}

	if msg.MinSelfDelegation != nil {
		if !msg.MinSelfDelegation.GT(validator.MinSelfDelegation) {
			return types.ErrMinSelfDelegationDecreased
		}

		if msg.MinSelfDelegation.GT(validator.Tokens) {
			return types.ErrSelfDelegationBelowMinimum
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
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress),
		),
	})

	return nil
}

// EpochDelegate logic is moved from msgServer.Delegate
func (k Keeper) EpochDelegate(ctx sdk.Context, msg *types.MsgDelegate) error {
	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return valErr
	}

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return sdkerrors.Wrapf(types.ErrBadDenom, "got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	// NOTE: source funds are always unbonded
	_, err = k.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonded, validator, true)
	if err != nil {
		return err
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
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return nil
}

// EpochBeginRedelegate logic is moved from msgServer.BeginRedelegate
func (k Keeper) EpochBeginRedelegate(ctx sdk.Context, msg *types.MsgBeginRedelegate) error {
	valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount,
	)
	if err != nil {
		return err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return sdkerrors.Wrapf(types.ErrBadDenom, "got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return err
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, shares,
	)
	if err != nil {
		return err
	}

	defer func() {
		telemetry.IncrCounter(1, types.ModuleName, "redelegate")
		telemetry.SetGaugeWithLabels(
			[]string{"tx", "msg", msg.Type()},
			float32(msg.Amount.Amount.Int64()),
			[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
		)
	}()

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return nil
}

// EpochUndelegate logic is moved from msgServer.Undelegate
func (k Keeper) EpochUndelegate(ctx sdk.Context, msg *types.MsgUndelegate) error {
	addr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, msg.Amount.Amount,
	)
	if err != nil {
		return err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return sdkerrors.Wrapf(types.ErrBadDenom, "got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	completionTime, err := k.Undelegate(ctx, delegatorAddress, addr, shares)
	if err != nil {
		return err
	}

	defer func() {
		telemetry.IncrCounter(1, types.ModuleName, "undelegate")
		telemetry.SetGaugeWithLabels(
			[]string{"tx", "msg", msg.Type()},
			float32(msg.Amount.Amount.Int64()),
			[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
		)
	}()

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return nil
}
