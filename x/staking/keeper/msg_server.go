package keeper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metrics "github.com/armon/go-metrics"
	tmstrings "github.com/tendermint/tendermint/libs/strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking/types"
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

	validator.MinSelfDelegation = msg.MinSelfDelegation

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.GetOperator())

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, msg.Value.Amount, sdkstaking.Unbonded, validator, true)
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

	// NOTE: source funds are always unbonded
	newShares, err := k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, sdkstaking.Unbonded, validator, true)
	if err != nil {
		return nil, err
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
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return nil, err
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, shares,
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
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
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

func getShareTokenDenom(validatorAddress string, tokenizeShareRecordId uint64) string {
	return validatorAddress + strconv.Itoa(int(tokenizeShareRecordId))
}

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

	_ = validator

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	delegation, found := k.GetDelegation(ctx, delegatorAddress, valAddr)
	if !found {
		return nil, types.ErrNoDelegatorForAddress
	}

	if msg.Amount.Denom != k.BondDenom(ctx) {
		return nil, types.ErrOnlyBondDenomAllowdForTokenize
	}

	delegationAmount := validator.Tokens.ToDec().Mul(delegation.GetShares()).Quo(validator.DelegatorShares)
	if msg.Amount.Amount.GT(sdk.Int(delegationAmount)) {
		return nil, types.ErrNotEnoughDelegationShares
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

	recordId := k.GetLastTokenizeShareRecordId(ctx) + 1
	k.SetLastTokenizeShareRecordId(ctx, recordId)

	shareTokenDenom := getShareTokenDenom(msg.ValidatorAddress, recordId)
	record := types.TokenizeShareRecord{
		Id:              recordId,
		Owner:           msg.TokenizedShareOwner,
		ShareTokenDenom: shareTokenDenom,
		ModuleAccount:   fmt.Sprintf("tokenizeshare_%d", recordId),
		Validator:       msg.ValidatorAddress,
	}

	shareToken := sdk.NewCoin(shareTokenDenom, msg.Amount.Amount)

	err = k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddress, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valAddr, msg.Amount.Amount,
	)

	returnAmount, err := k.Unbond(ctx, delegatorAddress, valAddr, shares)
	if err != nil {
		return nil, err
	}

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	// Note: UndelegateCoinsFromModuleToAccount is internally calling TrackUndelegation for vesting account
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, delegatorAddress, sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}

	// create reward ownership record
	k.AddTokenizeShareRecord(ctx, record)

	// send coins to module account
	err = k.bankKeeper.SendCoins(ctx, delegatorAddress, record.GetModuleAddress(), sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}

	validator, found = k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// delegate from module account
	_, err = k.Keeper.Delegate(ctx, record.GetModuleAddress(), msg.Amount.Amount, sdkstaking.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	return &types.MsgTokenizeSharesResponse{
		Amount: shareToken,
	}, nil
}

func (k msgServer) RedeemTokens(goCtx context.Context, msg *types.MsgRedeemTokensforShares) (*types.MsgRedeemTokensforSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	balance := k.bankKeeper.GetBalance(ctx, delegatorAddress, msg.Amount.Denom)
	if balance.Amount.LT(msg.Amount.Amount) {
		return nil, types.ErrNotEnoughBalance
	}

	record, err := k.GetTokenizeShareRecordByDenom(ctx, msg.Amount.Denom)
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

	// calculate the ratio between shares and redeem amount
	// moduleAccountTotalDelegation * redeemAmount / totalIssue
	delegation, found := k.GetDelegation(ctx, record.GetModuleAddress(), valAddr)
	shareDenomSupply := k.bankKeeper.GetSupply(ctx, msg.Amount.Denom)
	shares := delegation.Shares.Mul(msg.Amount.Amount.ToDec()).QuoInt(shareDenomSupply.Amount)

	returnAmount, err := k.Unbond(ctx, record.GetModuleAddress(), valAddr, shares)
	if err != nil {
		return nil, err
	}

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	_, found = k.GetDelegation(ctx, record.GetModuleAddress(), valAddr)
	if !found {
		if k.hooks != nil {
			k.hooks.BeforeTokenizeShareRecordRemoved(ctx, record.Id)
		}

		k.DeleteTokenizeShareRecord(ctx, record.Id)
	}

	// send share tokens to NotBondedPool and burn
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddress, types.NotBondedPoolName, sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.NotBondedPoolName, sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}

	// send equivalent amount of tokens to the delegator
	returnCoin := sdk.NewCoin(k.BondDenom(ctx), returnAmount)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, delegatorAddress, sdk.Coins{returnCoin})
	if err != nil {
		return nil, err
	}

	validator, found = k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// convert the share tokens to delegated status
	// Note: Delegate(substractAccount => true) -> DelegateCoinsFromAccountToModule -> TrackDelegation for vesting account
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, returnAmount, sdkstaking.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	return &types.MsgRedeemTokensforSharesResponse{}, nil
}

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
	k.deleteTokenizeShareRecordWithOwner(ctx, oldOwner, record.Id)

	record.Owner = msg.NewOwner
	k.setTokenizeShareRecord(ctx, record)

	// Set new account reference
	newOwner, err := sdk.AccAddressFromBech32(record.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	k.setTokenizeShareRecordWithOwner(ctx, newOwner, record.Id)

	return &types.MsgTransferTokenizeShareRecordResponse{}, nil
}
