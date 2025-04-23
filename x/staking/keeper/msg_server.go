package keeper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/go-metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the staking MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if err := msg.Validate(k.validatorAddressCodec); err != nil {
		return nil, err
	}

	minCommRate, err := k.MinCommissionRate(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Commission.Rate.LT(minCommRate) {
		return nil, errorsmod.Wrapf(types.ErrCommissionLTMinRate, "cannot set validator commission to less than minimum rate of %s", minCommRate)
	}

	// check to see if the pubkey or sender has been registered before
	if _, err := k.GetValidator(ctx, valAddr); err == nil {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, err := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); err == nil {
		return nil, types.ErrValidatorPubKeyExists
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Value.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Value.Denom, bondDenom,
		)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cp := sdkCtx.ConsensusParams()
	if cp.Validator != nil {
		pkType := pk.Type()
		hasKeyType := false
		for _, keyType := range cp.Validator.PubKeyTypes {
			if pkType == keyType {
				hasKeyType = true
				break
			}
		}
		if !hasKeyType {
			return nil, errorsmod.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := types.NewValidator(msg.ValidatorAddress, pk, msg.Description)
	if err != nil {
		return nil, err
	}

	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, sdkCtx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetValidatorByConsAddr(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetNewValidatorByPowerIndex(ctx, validator)
	if err != nil {
		return nil, err
	}

	// call the after-creation hook
	if err := k.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		return nil, err
	}

	// if this delegation is from a liquid staking provider (identified if the delegator
	// is an ICA account), it cannot exceed the global or validator bond cap
	if k.DelegatorIsLiquidStaker(valAddr) {
		shares, err := validator.SharesFromTokens(msg.Value.Amount)
		if err != nil {
			return nil, err
		}
		if err := k.SafelyIncreaseTotalLiquidStakedTokens(ctx, msg.Value.Amount, false); err != nil {
			return nil, err
		}
		validator, err = k.SafelyIncreaseValidatorLiquidShares(ctx, valAddr, shares, false)
		if err != nil {
			return nil, err
		}
	}

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Keeper.Delegate(ctx, sdk.AccAddress(valAddr), msg.Value.Amount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}

// EditValidator defines a method for editing an existing validator
func (k msgServer) EditValidator(ctx context.Context, msg *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Description == (types.Description{}) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(math.LegacyOneDec()) || msg.CommissionRate.IsNegative() {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commission rate must be between 0 and 1 (inclusive)")
		}

		minCommissionRate, err := k.MinCommissionRate(ctx)
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
		}

		if msg.CommissionRate.LT(minCommissionRate) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "commission rate cannot be less than the min commission rate %s", minCommissionRate.String())
		}
	}

	// validator must already be registered
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
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
		if err := k.Hooks().BeforeValidatorModified(ctx, valAddr); err != nil {
			return nil, err
		}

		validator.Commission = commission
	}

	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
		),
	})

	return &types.MsgEditValidatorResponse{}, nil
}

// Delegate defines a method for performing a delegation of coins from a delegator to a validator
func (k msgServer) Delegate(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	valAddr, valErr := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if valErr != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", valErr)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid delegation amount",
		)
	}

	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	tokens := msg.Amount.Amount

	// if this delegation is from a liquid staking provider (identified if the delegator
	// is an ICA account), it cannot exceed the global or validator bond cap
	if k.DelegatorIsLiquidStaker(delegatorAddress) {
		shares, err := validator.SharesFromTokens(tokens)
		if err != nil {
			return nil, err
		}
		if err := k.SafelyIncreaseTotalLiquidStakedTokens(ctx, tokens, false); err != nil {
			return nil, err
		}
		validator, err = k.SafelyIncreaseValidatorLiquidShares(ctx, valAddr, shares, false)
		if err != nil {
			return nil, err
		}
	}

	// NOTE: source funds are always unbonded
	newShares, err := k.Keeper.Delegate(ctx, delegatorAddress, tokens, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	// If the delegation is a validator bond, increment the validator bond shares
	delegation, err := k.Keeper.GetDelegation(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, err
	}
	if delegation.ValidatorBond {
		if err := k.IncreaseValidatorBondShares(ctx, valAddr, newShares); err != nil {
			return nil, err
		}
	}

	if tokens.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "delegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", sdk.MsgTypeURL(msg)},
				float32(tokens.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
		),
	})

	return &types.MsgDelegateResponse{}, nil
}

// BeginRedelegate defines a method for performing a redelegation of coins from a source validator to a destination validator of given delegator
func (k msgServer) BeginRedelegate(ctx context.Context, msg *types.MsgBeginRedelegate) (*types.MsgBeginRedelegateResponse, error) {
	valSrcAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid source validator address: %s", err)
	}
	valDstAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorDstAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid destination validator address: %s", err)
	}

	_, err = k.GetValidator(ctx, valSrcAddr)
	if err != nil {
		return nil, err
	}
	dstValidator, err := k.GetValidator(ctx, valDstAddr)
	if err != nil {
		return nil, err
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}
	srcDelegation, err := k.GetDelegation(ctx, delegatorAddress, valSrcAddr)
	if err != nil {
		return nil, sdkerrors.ErrNotFound.Wrapf(
			"delegation with delegator %s not found for validator %s. error: %s",
			msg.DelegatorAddress, msg.ValidatorSrcAddress, err)
	}

	srcShares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	// If this is a validator self-bond, the new liquid delegation cannot fall below the self-bond * bond factor
	// The delegation on the new validator will not be a validator bond
	if srcDelegation.ValidatorBond {
		if err := k.SafelyDecreaseValidatorBond(ctx, valSrcAddr, srcShares); err != nil {
			return nil, err
		}
	}

	// If this delegation from a liquid staker, the delegation on the new validator
	// cannot exceed that validator's self-bond cap
	// The liquid shares from the source validator should get moved to the destination validator
	if k.DelegatorIsLiquidStaker(delegatorAddress) {
		dstShares, err := dstValidator.SharesFromTokensTruncated(msg.Amount.Amount)
		if err != nil {
			return nil, err
		}
		if _, err := k.SafelyIncreaseValidatorLiquidShares(ctx, valDstAddr, dstShares, false); err != nil {
			return nil, err
		}
		if _, err := k.DecreaseValidatorLiquidShares(ctx, valSrcAddr, srcShares); err != nil {
			return nil, err
		}
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, srcShares,
	)
	if err != nil {
		return nil, err
	}

	// If the redelegation adds to a validator bond delegation, update the validator's bond shares
	dstDelegation, err := k.GetDelegation(ctx, delegatorAddress, valDstAddr)
	if err != nil {
		return nil, err
	}
	if dstDelegation.ValidatorBond {
		dstShares, err := dstValidator.SharesFromTokensTruncated(msg.Amount.Amount)
		if err != nil {
			return nil, err
		}
		if err := k.IncreaseValidatorBondShares(ctx, valDstAddr, dstShares); err != nil {
			return nil, err
		}
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "redelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", sdk.MsgTypeURL(msg)},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})

	return &types.MsgBeginRedelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

// Undelegate defines a method for performing an undelegation from a delegate and a validator
func (k msgServer) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	addr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	tokens := msg.Amount.Amount
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, tokens,
	)
	if err != nil {
		return nil, err
	}

	_, err = k.GetValidator(ctx, addr)
	if err != nil {
		return nil, err
	}

	delegation, err := k.GetDelegation(ctx, delegatorAddress, addr)
	if err != nil {
		return nil, sdkerrors.ErrNotFound.Wrapf(
			"delegation with delegator %s not found for validator %s. error: %s",
			msg.DelegatorAddress, msg.ValidatorAddress, err)
	}

	// if this is a validator self-bond, the new liquid delegation cannot fall below the self-bond * bond factor
	if delegation.ValidatorBond {
		if err := k.SafelyDecreaseValidatorBond(ctx, addr, shares); err != nil {
			return nil, err
		}
	}

	// if this delegation is from a liquid staking provider (identified if the delegator
	// is an ICA account), the global and validator liquid totals should be decremented
	if k.DelegatorIsLiquidStaker(delegatorAddress) {
		if err := k.DecreaseTotalLiquidStakedTokens(ctx, tokens); err != nil {
			return nil, err
		}
		if _, err := k.DecreaseValidatorLiquidShares(ctx, addr, shares); err != nil {
			return nil, err
		}
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	completionTime, undelegatedAmt, err := k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
	if err != nil {
		return nil, err
	}

	undelegatedCoin := sdk.NewCoin(msg.Amount.Denom, undelegatedAmt)

	if tokens.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "undelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", sdk.MsgTypeURL(msg)},
				float32(tokens.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, undelegatedCoin.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})

	return &types.MsgUndelegateResponse{
		CompletionTime: completionTime,
		Amount:         undelegatedCoin,
	}, nil
}

// CancelUnbondingDelegation defines a method for canceling the unbonding delegation
// and delegate back to the validator.
func (k msgServer) CancelUnbondingDelegation(ctx context.Context, msg *types.MsgCancelUnbondingDelegation) (*types.MsgCancelUnbondingDelegationResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid amount",
		)
	}

	if msg.CreationHeight <= 0 {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid height",
		)
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// In some situations, the exchange rate becomes invalid, e.g. if
	// Validator loses all tokens due to slashing. In this case,
	// make all future delegations invalid.
	if validator.InvalidExRate() {
		return nil, types.ErrDelegatorShareExRateInvalid
	}

	if validator.IsJailed() {
		return nil, types.ErrValidatorJailed
	}

	ubd, err := k.GetUnbondingDelegation(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"unbonding delegation with delegator %s not found for validator %s",
			msg.DelegatorAddress, msg.ValidatorAddress,
		)
	}

	// if this undelegation was from a liquid staking provider (identified if the delegator
	// is an ICA account), the global and validator liquid totals should be incremented
	tokens := msg.Amount.Amount
	if k.DelegatorIsLiquidStaker(delegatorAddress) {
		shares, err := validator.SharesFromTokens(tokens)
		if err != nil {
			return nil, err
		}
		if err := k.SafelyIncreaseTotalLiquidStakedTokens(ctx, tokens, false); err != nil {
			return nil, err
		}
		validator, err = k.SafelyIncreaseValidatorLiquidShares(ctx, valAddr, shares, false)
		if err != nil {
			return nil, err
		}
	}

	var (
		unbondEntry      types.UnbondingDelegationEntry
		unbondEntryIndex int64 = -1
	)

	for i, entry := range ubd.Entries {
		if entry.CreationHeight == msg.CreationHeight {
			unbondEntry = entry
			unbondEntryIndex = int64(i)
			break
		}
	}
	if unbondEntryIndex == -1 {
		return nil, sdkerrors.ErrNotFound.Wrapf("unbonding delegation entry is not found at block height %d", msg.CreationHeight)
	}

	if unbondEntry.Balance.LT(msg.Amount.Amount) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("amount is greater than the unbonding delegation entry balance")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if unbondEntry.CompletionTime.Before(sdkCtx.BlockTime()) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("unbonding delegation is already processed")
	}

	// delegate back the unbonding delegation amount to the validator
	newShares, err := k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonding, validator, false)
	if err != nil {
		return nil, err
	}

	// If the delegation is a validator bond, increment the validator bond shares
	delegation, err := k.Keeper.GetDelegation(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, err
	}
	if delegation.ValidatorBond {
		if err := k.IncreaseValidatorBondShares(ctx, valAddr, newShares); err != nil {
			return nil, err
		}
	}

	amount := unbondEntry.Balance.Sub(msg.Amount.Amount)
	if amount.IsZero() {
		ubd.RemoveEntry(unbondEntryIndex)
	} else {
		// update the unbondingDelegationEntryBalance and InitialBalance for ubd entry
		unbondEntry.Balance = amount
		unbondEntry.InitialBalance = unbondEntry.InitialBalance.Sub(msg.Amount.Amount)
		ubd.Entries[unbondEntryIndex] = unbondEntry
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		err = k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		err = k.SetUnbondingDelegation(ctx, ubd)
	}

	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCancelUnbondingDelegation,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyCreationHeight, strconv.FormatInt(msg.CreationHeight, 10)),
		),
	)

	return &types.MsgCancelUnbondingDelegationResponse{}, nil
}

// UpdateParams defines a method to perform updation of params exist in x/staking module.
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// store params
	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// UnbondValidator defines a method for performing the status transition for
// a validator from bonded to unbonding
// This allows a validator to stop their services and jail themselves without
// experiencing a slash
func (k msgServer) UnbondValidator(ctx context.Context, msg *types.MsgUnbondValidator) (*types.MsgUnbondValidatorResponse, error) {
	// convert sdk.AccAddress to sdk.ValAddress
	accAddr, err := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	valAddr, err := k.validatorAddressCodec.StringToBytes(sdk.ValAddress(accAddr).String())
	if err != nil {
		return nil, err
	}

	// validator must already be registered
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	err = k.jailValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnbondValidatorResponse{}, nil
}

/////////////////////////////////////// DISABLED LSM CODE ///////////////////////////////////////

// Tokenizes shares associated with a delegation by creating a tokenize share record
// and returning tokens with a denom of the format {validatorAddress}/{recordId}
func (k msgServer) TokenizeShares(_ context.Context, _ *types.MsgTokenizeShares) (*types.MsgTokenizeSharesResponse, error) {
	return &types.MsgTokenizeSharesResponse{}, fmt.Errorf("tokenize shares is disabled - use x/liquid")
}

func (k msgServer) TransferTokenizeShareRecord(_ context.Context, _ *types.MsgTransferTokenizeShareRecord) (*types.MsgTransferTokenizeShareRecordResponse, error) {
	return &types.MsgTransferTokenizeShareRecordResponse{}, fmt.Errorf("transfer tokens for shares is disabled - use x/liquid")

}

// Converts tokenized shares back into a native delegation
func (k msgServer) RedeemTokensForShares(v context.Context, _ *types.MsgRedeemTokensForShares) (*types.MsgRedeemTokensForSharesResponse, error) {
	return &types.MsgRedeemTokensForSharesResponse{}, fmt.Errorf("redeem tokens for shares is disabled - use x/liquid")
}

// DisableTokenizeShares prevents an address from tokenizing any of their delegations
func (k msgServer) DisableTokenizeShares(_ context.Context, _ *types.MsgDisableTokenizeShares) (*types.MsgDisableTokenizeSharesResponse, error) {
	return &types.MsgDisableTokenizeSharesResponse{}, fmt.Errorf("disable tokenize shares is disabled - use x/liquid")
}

// EnableTokenizeShares begins the countdown after which tokenizing shares by the
// sender address is re-allowed, which will complete after the unbonding period
func (k msgServer) EnableTokenizeShares(_ context.Context, _ *types.MsgEnableTokenizeShares) (*types.MsgEnableTokenizeSharesResponse, error) {
	return &types.MsgEnableTokenizeSharesResponse{}, fmt.Errorf("enable tokenize shares is disabled - use x/liquid")
}

// Designates a delegation as a validator bond
// This enables the validator to receive more liquid staking delegations
func (k msgServer) ValidatorBond(_ context.Context, _ *types.MsgValidatorBond) (*types.MsgValidatorBondResponse, error) {
	return &types.MsgValidatorBondResponse{}, fmt.Errorf("validator bond is disabled - use x/liquid")
}
