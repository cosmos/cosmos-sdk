package keeper

import (
	"fmt"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ExecuteQueuedCreateValidatorMsg handles the execution of a queued MsgCreateValidator.
// The validator has already been created at this point, so all that remains is the self delegation
func (k Keeper) executeQueuedCreateValidatorMsg(ctx sdk.Context, msg *types.MsgCreateValidator) error {
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

// revertCreateValidatorMsg does cancel self-delegation
func (k Keeper) revertCreateValidatorMsg(ctx sdk.Context, msg *types.MsgCreateValidator) error {

	bondDenom := k.BondDenom(ctx)
	if msg.Value.Denom != bondDenom {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", msg.Value.Denom, bondDenom)
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), msg.Value.Amount))
	if err := k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, types.EpochDelegationPoolName, delegatorAddress, coins); err != nil {
		return err
	}

	// TODO: we might need to delete validator if validator's delegation amount is zero
	// TODO: report somewhere for logging cancel event
	return nil
}

// executeQueuedEditValidatorMsg logic is moved from msgServer.EditValidator
func (k Keeper) executeQueuedEditValidatorMsg(ctx sdk.Context, msg *types.MsgEditValidator) error {
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

// executeQueuedDelegationMsg logic is moved from msgServer.Delegate
func (k Keeper) executeQueuedDelegationMsg(ctx sdk.Context, msg *types.MsgDelegate) error {
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
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	// NOTE: source funds are always unbonded
	_, err = k.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonded, validator, true)
	if err != nil {
		return err
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "delegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

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

// revertDelegationMsg does cancel delegation
func (k Keeper) revertDelegationMsg(ctx sdk.Context, msg *types.MsgDelegate) error {
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	coins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), msg.Amount.Amount))
	if err := k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, types.EpochDelegationPoolName, delegatorAddress, coins); err != nil {
		return err
	}

	// TODO: report somewhere for logging cancel event
	return nil
}

// executeQueuedBeginRedelegateMsg logic is moved from msgServer.BeginRedelegate
func (k Keeper) executeQueuedBeginRedelegateMsg(ctx sdk.Context, msg *types.MsgBeginRedelegate) (time.Time, error) {
	valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return time.Time{}, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return time.Time{}, err
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount,
	)
	if err != nil {
		return time.Time{}, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return time.Time{}, sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return time.Time{}, err
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, shares,
	)
	if err != nil {
		return time.Time{}, err
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "redelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

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

	return completionTime, nil
}

// executeQueuedUndelegateMsg logic is moved from msgServer.Undelegate
func (k Keeper) executeQueuedUndelegateMsg(ctx sdk.Context, msg *types.MsgUndelegate) (time.Time, error) {
	addr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return time.Time{}, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return time.Time{}, err
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, msg.Amount.Amount,
	)
	if err != nil {
		return time.Time{}, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return time.Time{}, sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom)
	}

	completionTime, err := k.Undelegate(ctx, delegatorAddress, addr, shares)
	if err != nil {
		return time.Time{}, err
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "undelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

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

	return completionTime, nil
}

// ExecuteEpoch execute epoch actions
// If a message fails to execute, the message will be reverted if the tokens were taken from an account,
// otherwise an error log will be produced.
// Unfortunately, the user will need to check if all there messages were executed.
func (k Keeper) ExecuteEpoch(ctx sdk.Context) {
	cacheCtx, writeCache := ctx.CacheContext()
	logger := k.Logger(ctx)

	// execute all epoch actions
	for iterator := k.epochKeeper.GetEpochActionsIterator(ctx); iterator.Valid(); iterator.Next() {
		msg := k.GetEpochActionByIterator(iterator)

		switch msg := msg.(type) {
		case *types.MsgCreateValidator:
			err := k.executeQueuedCreateValidatorMsg(cacheCtx, msg)
			if err == nil {
				writeCache()
			} else if err = k.revertCreateValidatorMsg(ctx, msg); err != nil {
				// avoid panicking
				logger.Error("create validator failed to execute", "msg", msg, "err", err)
			}
		case *types.MsgEditValidator:
			err := k.executeQueuedEditValidatorMsg(cacheCtx, msg)
			if err == nil {
				writeCache()
			} else {
				// TODO: report somewhere for logging edit not success or panic
				logger.Error("edit validator failed to execute", "msg", msg, "err", err)
				// panic(fmt.Sprintf("not be able to execute, %T", msg))
			}
		case *types.MsgDelegate:
			err := k.executeQueuedDelegationMsg(cacheCtx, msg)
			if err == nil {
				writeCache()
			} else if err = k.revertDelegationMsg(ctx, msg); err != nil {
				logger.Error("not be able to execute nor revert MsgDelegate", "msg", msg, "err", err)
			}
		case *types.MsgBeginRedelegate:
			_, err := k.executeQueuedBeginRedelegateMsg(cacheCtx, msg)
			if err == nil {
				writeCache()
			} else {
				logger.Error("edit begin redelegation failed to execute", "msg", msg, "err", err)
			}
		case *types.MsgUndelegate:
			_, err := k.executeQueuedUndelegateMsg(cacheCtx, msg)
			if err == nil {
				writeCache()
			} else {
				logger.Error("edit undelegation failed to execute", "msg", msg, "err", err)
			}
		default:
			panic(fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
		// dequeue processed item
		k.epochKeeper.DeleteByKey(ctx, iterator.Key())
	}

	// Update epochNumber after epoch finish
	// This won't affect slashing module since slashing Endblocker run before staking module
	k.epochKeeper.IncreaseEpochNumber(ctx)
}
