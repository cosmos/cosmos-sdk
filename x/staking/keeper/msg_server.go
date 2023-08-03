package keeper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metrics "github.com/armon/go-metrics"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(goCtx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, valAddr); found {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); found {
		return nil, types.ErrValidatorPubKeyExists
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Value.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Value.Denom, bondDenom,
		)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	cp := ctx.ConsensusParams()
	if cp != nil && cp.Validator != nil {
		if !tmstrings.StringInSlice(pk.Type(), cp.Validator.PubKeyTypes) {
			return nil, sdkerrors.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := types.NewValidator(valAddr, pk, msg.Description)
	if err != nil {
		return nil, err
	}
	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	k.SetValidator(ctx, validator)
	err = k.SetValidatorByConsAddr(ctx, validator)
	if err != nil {
		return nil, err
	}
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.GetOperator())

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, msg.Value.Amount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}

// EditValidator defines a method for editing an existing validator
func (k msgServer) EditValidator(goCtx context.Context, msg *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	// validator must already be registered
	validator, found := k.GetValidator(ctx, valAddr)
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
		k.BeforeValidatorModified(ctx, valAddr)

		validator.Commission = commission
	}

	k.SetValidator(ctx, validator)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress),
		),
	})

	return &types.MsgEditValidatorResponse{}, nil
}

// Delegate defines a method for performing a delegation of coins from a delegator to a validator
func (k msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
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
		if err := k.SafelyIncreaseValidatorLiquidShares(ctx, validator, shares); err != nil {
			return nil, err
		}
	}

	// Note: this is required for downstream uses of the validator variable
	// since the validator's liquid shares were updated above
	validator, found = k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// NOTE: source funds are always unbonded
	newShares, err := k.Keeper.Delegate(ctx, delegatorAddress, tokens, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	// If the delegation is a validator bond, increment the validator bond shares
	delegation, found := k.Keeper.GetDelegation(ctx, delegatorAddress, valAddr)
	if !found {
		return nil, types.ErrNoDelegation
	}
	if delegation.ValidatorBond {
		k.IncreaseValidatorBondShares(ctx, validator, newShares)
	}

	if tokens.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "delegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(tokens.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgDelegateResponse{}, nil
}

// BeginRedelegate defines a method for performing a redelegation of coins from a delegator and source validator to a destination validator
func (k msgServer) BeginRedelegate(goCtx context.Context, msg *types.MsgBeginRedelegate) (*types.MsgBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, err
	}
	valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return nil, err
	}

	srcValidator, found := k.GetValidator(ctx, valSrcAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}
	dstValidator, found := k.GetValidator(ctx, valDstAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	delegation, found := k.GetDelegation(ctx, delegatorAddress, valSrcAddr)
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"delegation with delegator %s not found for validator %s",
			msg.DelegatorAddress, msg.ValidatorSrcAddress,
		)
	}

	srcShares, err := k.ValidateUnbondAmount(ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount)
	if err != nil {
		return nil, err
	}

	// If this is a validator self-bond, the new liquid delegation cannot fall below the self-bond * bond factor
	// The delegation on the new validator will not a validator bond
	if delegation.ValidatorBond {
		if err := k.SafelyDecreaseValidatorBond(ctx, srcValidator, srcShares); err != nil {
			return nil, err
		}
		// Note: this is required for downstream uses of the srcValidator variable
		// since the validator bond shares were updated above
		srcValidator, found = k.GetValidator(ctx, valSrcAddr)
		if !found {
			return nil, types.ErrNoValidatorFound
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
		if err := k.SafelyIncreaseValidatorLiquidShares(ctx, dstValidator, dstShares); err != nil {
			return nil, err
		}
		if err := k.DecreaseValidatorLiquidShares(ctx, srcValidator, srcShares); err != nil {
			return nil, err
		}
		// Note: this is required for downstream uses of each validator variable
		// since the liquid shares were updated above
		srcValidator, found = k.GetValidator(ctx, valSrcAddr)
		if !found {
			return nil, types.ErrNoValidatorFound
		}
		dstValidator, found = k.GetValidator(ctx, valDstAddr)
		if !found {
			return nil, types.ErrNoValidatorFound
		}
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, srcShares,
	)
	if err != nil {
		return nil, err
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
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgBeginRedelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

// Undelegate defines a method for performing an undelegation from a delegate and a validator
func (k msgServer) Undelegate(goCtx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	tokens := msg.Amount.Amount
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, tokens,
	)
	if err != nil {
		return nil, err
	}

	validator, found := k.GetValidator(ctx, addr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	delegation, found := k.GetDelegation(ctx, delegatorAddress, addr)
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"delegation with delegator %s not found for validator %s",
			msg.DelegatorAddress, msg.ValidatorAddress,
		)
	}

	// if this is a validator self-bond, the new liquid delegation cannot fall below the self-bond * bond factor
	if delegation.ValidatorBond {
		if err := k.SafelyDecreaseValidatorBond(ctx, validator, shares); err != nil {
			return nil, err
		}
		// Note: this is required for downstream uses of the validator variable
		// since the validator bond shares was updated above
		validator, found = k.GetValidator(ctx, addr)
		if !found {
			return nil, types.ErrNoValidatorFound
		}
	}

	// if this delegation is from a liquid staking provider (identified if the delegator
	// is an ICA account), the global and validator liquid totals should be decremented
	if k.DelegatorIsLiquidStaker(delegatorAddress) {
		if err := k.DecreaseTotalLiquidStakedTokens(ctx, tokens); err != nil {
			return nil, err
		}
		if err := k.DecreaseValidatorLiquidShares(ctx, validator, shares); err != nil {
			return nil, err
		}
		// Note: this is required for downstream uses of the validator variable
		// since the liquid shares were updated above
		validator, found = k.GetValidator(ctx, addr)
		if !found {
			return nil, types.ErrNoValidatorFound
		}
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	completionTime, err := k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
	if err != nil {
		return nil, err
	}

	if tokens.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "undelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(tokens.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgUndelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

// CancelUnbondingDelegation defines a method for canceling the unbonding delegation
// and delegate back to the validator.
//
// This has been backported from SDK 46 as a desirable safety feature for LSM.
// If a liquid staking provider is exploited and the exploiter initiates an undelegation,
// having access to CancelUnbondingDelegation allows the liquid staking provider to cancel
// the undelegation with a software upgrade and thus avoid loss of user funds
func (k msgServer) CancelUnbondingDelegation(goCtx context.Context, msg *types.MsgCancelUnbondingDelegation) (*types.MsgCancelUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
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

	ubd, found := k.GetUnbondingDelegation(ctx, delegatorAddress, valAddr)
	if !found {
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
		if err := k.SafelyIncreaseValidatorLiquidShares(ctx, validator, shares); err != nil {
			return nil, err
		}
		// Note: this is required for downstream uses of the validator variable
		// since the validator's liquid shares were updated above
		validator, found = k.GetValidator(ctx, valAddr)
		if !found {
			return nil, types.ErrNoValidatorFound
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

	if unbondEntry.CompletionTime.Before(ctx.BlockTime()) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("unbonding delegation is already processed")
	}

	// delegate back the unbonding delegation amount to the validator
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonding, validator, false)
	if err != nil {
		return nil, err
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
		k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		k.SetUnbondingDelegation(ctx, ubd)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_unbonding_delegation",
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute("creation_height", strconv.FormatInt(msg.CreationHeight, 10)),
		),
	)

	return &types.MsgCancelUnbondingDelegationResponse{}, nil
}

// UnbondValidator defines a method for performing the status transition for
// a validator from bonded to unbonding
// This allows a validator to stop their services and jail themselves without
// experiencing a slash
func (k msgServer) UnbondValidator(goCtx context.Context, msg *types.MsgUnbondValidator) (*types.MsgUnbondValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	// validator must already be registered
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// jail the validator.
	k.jailValidator(ctx, validator)
	return &types.MsgUnbondValidatorResponse{}, nil
}

// Tokenizes shares associated with a delegation by creating a tokenize share record
// and returning tokens with a denom of the format {validatorAddress}/{recordId}
func (k msgServer) TokenizeShares(goCtx context.Context, msg *types.MsgTokenizeShares) (*types.MsgTokenizeSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// Check if the delegator has disabled tokenization
	lockStatus, unlockTime := k.GetTokenizeSharesLock(ctx, delegatorAddress)
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED {
		return nil, types.ErrTokenizeSharesDisabledForAccount
	}
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING {
		return nil, types.ErrTokenizeSharesDisabledForAccount.Wrapf("tokenization will be allowed at %s", unlockTime)
	}

	delegation, found := k.GetDelegation(ctx, delegatorAddress, valAddr)
	if !found {
		return nil, types.ErrNoDelegatorForAddress
	}

	if delegation.ValidatorBond {
		return nil, types.ErrValidatorBondNotAllowedForTokenizeShare
	}

	if msg.Amount.Denom != k.BondDenom(ctx) {
		return nil, types.ErrOnlyBondDenomAllowdForTokenize
	}

	acc := k.authKeeper.GetAccount(ctx, delegatorAddress)
	if acc != nil {
		acc, ok := acc.(vesting.VestingAccount)
		if ok {
			// if account is a vesting account, it checks if free delegation (non-vesting delegation) is not exceeding
			// the tokenize share amount and execute further tokenize share process
			// tokenize share is reducing unlocked tokens delegation from the vesting account and further process
			// is not causing issues
			delFree := acc.GetDelegatedFree().AmountOf(msg.Amount.Denom)
			if delFree.LT(msg.Amount.Amount) {
				return nil, types.ErrExceedingFreeVestingDelegations
			}
		}
	}

	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valAddr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	// If this tokenization is NOT from a liquid staking provider,
	//   confirm it does not exceed the global and validator liquid staking cap
	// If the tokenization is from a liquid staking provider,
	//   the shares are already considered liquid and there's no need to increment the totals
	if !k.DelegatorIsLiquidStaker(delegatorAddress) {
		if err := k.SafelyIncreaseTotalLiquidStakedTokens(ctx, msg.Amount.Amount, true); err != nil {
			return nil, err
		}
		if err := k.SafelyIncreaseValidatorLiquidShares(ctx, validator, shares); err != nil {
			return nil, err
		}
		// Note: this is required for downstream uses of the validator variable
		// since the validator's liquid shares were updated above
		validator, found = k.GetValidator(ctx, valAddr)
		if !found {
			return nil, types.ErrNoValidatorFound
		}
	}

	recordID := k.GetLastTokenizeShareRecordID(ctx) + 1
	k.SetLastTokenizeShareRecordID(ctx, recordID)

	record := types.TokenizeShareRecord{
		Id:            recordID,
		Owner:         msg.TokenizedShareOwner,
		ModuleAccount: fmt.Sprintf("%s%d", types.TokenizeShareModuleAccountPrefix, recordID),
		Validator:     msg.ValidatorAddress,
	}

	// note: this returnAmount can be slightly off from the original delegation amount if there
	// is a decimal to int precision error
	returnAmount, err := k.Unbond(ctx, delegatorAddress, valAddr, shares)
	if err != nil {
		return nil, err
	}

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	// Note: UndelegateCoinsFromModuleToAccount is internally calling TrackUndelegation for vesting account
	returnCoin := sdk.NewCoin(k.BondDenom(ctx), returnAmount)
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, delegatorAddress, sdk.Coins{returnCoin})
	if err != nil {
		return nil, err
	}

	// Re-calculate the shares in case there was rounding precision during the undelegation
	newShares, err := validator.SharesFromTokens(returnAmount)
	if err != nil {
		return nil, err
	}

	// The share tokens returned maps 1:1 with shares
	shareToken := sdk.NewCoin(record.GetShareTokenDenom(), newShares.TruncateInt())

	err = k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddress, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	// create reward ownership record
	err = k.AddTokenizeShareRecord(ctx, record)
	if err != nil {
		return nil, err
	}
	// send coins to module account
	err = k.bankKeeper.SendCoins(ctx, delegatorAddress, record.GetModuleAddress(), sdk.Coins{returnCoin})
	if err != nil {
		return nil, err
	}

	// Note: it is needed to get latest validator object to get Keeper.Delegate function work properly
	validator, found = k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// delegate from module account
	_, err = k.Keeper.Delegate(ctx, record.GetModuleAddress(), returnAmount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenizeShares,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyShareOwner, msg.TokenizedShareOwner),
			sdk.NewAttribute(types.AttributeKeyShareRecordID, fmt.Sprintf("%d", record.Id)),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
		),
	)

	return &types.MsgTokenizeSharesResponse{
		Amount: shareToken,
	}, nil
}

// Converts tokenized shares back into a native delegation
func (k msgServer) RedeemTokensForShares(goCtx context.Context, msg *types.MsgRedeemTokensForShares) (*types.MsgRedeemTokensForSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	shareToken := msg.Amount
	balance := k.bankKeeper.GetBalance(ctx, delegatorAddress, shareToken.Denom)
	if balance.Amount.LT(shareToken.Amount) {
		return nil, types.ErrNotEnoughBalance
	}

	record, err := k.GetTokenizeShareRecordByDenom(ctx, shareToken.Denom)
	if err != nil {
		return nil, err
	}

	valAddr, valErr := sdk.ValAddressFromBech32(record.Validator)
	if valErr != nil {
		return nil, valErr
	}

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	delegation, found := k.GetDelegation(ctx, record.GetModuleAddress(), valAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	// Similar to undelegations, if the account is attempting to tokenize the full delegation,
	// but there's a precision error due to the decimal to int conversion, round up to the
	// full decimal amount before modifying the delegation
	shares := shareToken.Amount.ToDec()
	if shareToken.Amount.Equal(delegation.Shares.TruncateInt()) {
		shares = delegation.Shares
	}
	tokens := validator.TokensFromShares(shares).TruncateInt()

	// If this redemption is NOT from a liquid staking provider, decrement the total liquid staked
	// If the redemption was from a liquid staking provider, the shares are still considered
	// liquid, even in their non-tokenized form (since they are owned by a liquid staking provider)
	if !k.DelegatorIsLiquidStaker(delegatorAddress) {
		if err := k.DecreaseTotalLiquidStakedTokens(ctx, tokens); err != nil {
			return nil, err
		}
		if err := k.DecreaseValidatorLiquidShares(ctx, validator, shares); err != nil {
			return nil, err
		}
		// Note: this is required for downstream uses of the validator variable
		// since the liquid shares were updated above
		validator, found = k.GetValidator(ctx, valAddr)
		if !found {
			return nil, types.ErrNoValidatorFound
		}
	}

	returnAmount, err := k.Unbond(ctx, record.GetModuleAddress(), valAddr, shares)
	if err != nil {
		return nil, err
	}

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	// Note: since delegation object has been changed from unbond call, it gets latest delegation
	_, found = k.GetDelegation(ctx, record.GetModuleAddress(), valAddr)
	if !found {
		if k.hooks != nil {
			if err := k.hooks.BeforeTokenizeShareRecordRemoved(ctx, record.Id); err != nil {
				return nil, err
			}
		}
		err = k.DeleteTokenizeShareRecord(ctx, record.Id)
		if err != nil {
			return nil, err
		}
	}

	// send share tokens to NotBondedPool and burn
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddress, types.NotBondedPoolName, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.NotBondedPoolName, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	// send equivalent amount of tokens to the delegator
	returnCoin := sdk.NewCoin(k.BondDenom(ctx), returnAmount)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, delegatorAddress, sdk.Coins{returnCoin})
	if err != nil {
		return nil, err
	}

	// Note: it is needed to get latest validator object to get Keeper.Delegate function work properly
	validator, found = k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// convert the share tokens to delegated status
	// Note: Delegate(substractAccount => true) -> DelegateCoinsFromAccountToModule -> TrackDelegation for vesting account
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, returnAmount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRedeemShares,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyValidator, validator.OperatorAddress),
			sdk.NewAttribute(types.AttributeKeyAmount, shareToken.String()),
		),
	)

	return &types.MsgRedeemTokensForSharesResponse{
		Amount: returnCoin,
	}, nil
}

// Transfers the ownership of rewards associated with a tokenize share record
func (k msgServer) TransferTokenizeShareRecord(goCtx context.Context, msg *types.MsgTransferTokenizeShareRecord) (*types.MsgTransferTokenizeShareRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	record, err := k.GetTokenizeShareRecord(ctx, msg.TokenizeShareRecordId)
	if err != nil {
		return nil, types.ErrTokenizeShareRecordNotExists
	}

	if record.Owner != msg.Sender {
		return nil, types.ErrNotTokenizeShareRecordOwner
	}

	// Remove old account reference
	oldOwner, err := sdk.AccAddressFromBech32(record.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	k.deleteTokenizeShareRecordWithOwner(ctx, oldOwner, record.Id)

	record.Owner = msg.NewOwner
	k.setTokenizeShareRecord(ctx, record)

	// Set new account reference
	newOwner, err := sdk.AccAddressFromBech32(record.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	k.setTokenizeShareRecordWithOwner(ctx, newOwner, record.Id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransferTokenizeShareRecord,
			sdk.NewAttribute(types.AttributeKeyShareRecordID, fmt.Sprintf("%d", msg.TokenizeShareRecordId)),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyShareOwner, msg.NewOwner),
		),
	)

	return &types.MsgTransferTokenizeShareRecordResponse{}, nil
}

// DisableTokenizeShares prevents an address from tokenizing any of their delegations
func (k msgServer) DisableTokenizeShares(goCtx context.Context, msg *types.MsgDisableTokenizeShares) (*types.MsgDisableTokenizeSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegator := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)

	// If tokenized shares is already disabled, alert the user
	lockStatus, completionTime := k.GetTokenizeSharesLock(ctx, delegator)
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED {
		return nil, types.ErrTokenizeSharesAlreadyDisabledForAccount
	}

	// If the tokenized shares lock is expiring, remove the pending unlock from the queue
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING {
		k.CancelTokenizeShareLockExpiration(ctx, delegator, completionTime)
	}

	// Create a new tokenization lock for the user
	// Note: if there is a lock expiration in progress, this will override the expiration
	k.AddTokenizeSharesLock(ctx, delegator)

	return &types.MsgDisableTokenizeSharesResponse{}, nil
}

// EnableTokenizeShares begins the countdown after which tokenizing shares by the
// sender address is re-allowed, which will complete after the unbonding period
func (k msgServer) EnableTokenizeShares(goCtx context.Context, msg *types.MsgEnableTokenizeShares) (*types.MsgEnableTokenizeSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegator := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)

	// If tokenized shares aren't current disabled, alert the user
	lockStatus, unlockTime := k.GetTokenizeSharesLock(ctx, delegator)
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED {
		return nil, types.ErrTokenizeSharesAlreadyEnabledForAccount
	}
	if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING {
		return nil, types.ErrTokenizeSharesAlreadyEnabledForAccount.Wrapf(
			"tokenize shares re-enablement already in progress, ending at %s", unlockTime)
	}

	// Otherwise queue the unlock
	completionTime := k.QueueTokenizeSharesAuthorization(ctx, delegator)

	return &types.MsgEnableTokenizeSharesResponse{CompletionTime: completionTime}, nil
}

// Designates a delegation as a validator bond
// This enables the validator to receive more liquid staking delegations
func (k msgServer) ValidatorBond(goCtx context.Context, msg *types.MsgValidatorBond) (*types.MsgValidatorBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	delegation, found := k.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil, types.ErrNoDelegation
	}

	// liquid staking providers should not be able to validator bond
	if k.DelegatorIsLiquidStaker(delAddr) {
		return nil, types.ErrValidatorBondNotAllowedFromModuleAccount
	}

	if !delegation.ValidatorBond {
		delegation.ValidatorBond = true
		k.SetDelegation(ctx, delegation)
		validator.ValidatorBondShares = validator.ValidatorBondShares.Add(delegation.Shares)
		k.SetValidator(ctx, validator)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeValidatorBondDelegation,
				sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			),
		)
	}

	return &types.MsgValidatorBondResponse{}, nil
}
