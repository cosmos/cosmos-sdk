package keeper

import (
	"context"
	goerrors "errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Keeper of the distribution store
type Keeper struct {
	storeService  store.KVStoreService
	cdc           codec.BinaryCodec
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	Schema  collections.Schema
	Params  collections.Item[types.Params]
	FeePool collections.Item[types.FeePool]

	feeCollectorName string // name of the FeeCollector ModuleAccount
}

// NewKeeper creates a new distribution Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService,
	ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper,
	feeCollectorName, authority string,
) Keeper {
	// ensure distribution module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		storeService:     storeService,
		cdc:              cdc,
		authKeeper:       ak,
		bankKeeper:       bk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
		Params:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		FeePool:          collections.NewItem(sb, types.FeePoolKey, "fee_pool", codec.CollValue[types.FeePool](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the x/distribution module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With(log.ModuleKey, "x/"+types.ModuleName)
}

// SetWithdrawAddr sets a new address that will receive the rewards upon withdrawal
func (k Keeper) SetWithdrawAddr(ctx context.Context, delegatorAddr, withdrawAddr sdk.AccAddress) error {
	if k.bankKeeper.BlockedAddr(withdrawAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive external funds", withdrawAddr)
	}

	withdrawAddrEnabled, err := k.GetWithdrawAddrEnabled(ctx)
	if err != nil {
		return err
	}

	if !withdrawAddrEnabled {
		return types.ErrSetWithdrawAddrDisabled
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetWithdrawAddress,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, withdrawAddr.String()),
		),
	)

	k.SetDelegatorWithdrawAddr(ctx, delegatorAddr, withdrawAddr)
	return nil
}

// withdraw rewards from a delegation
func (k Keeper) WithdrawDelegationRewards(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	val, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, types.ErrNoValidatorDistInfo
	}

	del, err := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	if err != nil {
		return nil, err
	}

	if del == nil {
		return nil, types.ErrEmptyDelegationDistInfo
	}

	// withdraw rewards
	rewards, err := k.withdrawDelegationRewards(ctx, val, del)
	if err != nil {
		return nil, err
	}

	// reinitialize the delegation
	err = k.initializeDelegation(ctx, valAddr, delAddr)
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

// withdraw validator commission
func (k Keeper) WithdrawValidatorCommission(ctx context.Context, valAddr sdk.ValAddress) (sdk.Coins, error) {
	// fetch validator accumulated commission
	accumCommission, err := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	if accumCommission.Commission.IsZero() {
		return nil, types.ErrNoValidatorCommission
	}

	commission, remainder := accumCommission.Commission.TruncateDecimal()
	k.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: remainder}) // leave remainder to withdraw later

	// update outstanding
	outstanding, err := k.GetValidatorOutstandingRewards(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	err = k.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: outstanding.Rewards.Sub(sdk.NewDecCoinsFromCoins(commission...))})
	if err != nil {
		return nil, err
	}

	if !commission.IsZero() {
		accAddr := sdk.AccAddress(valAddr)
		withdrawAddr, err := k.GetDelegatorWithdrawAddr(ctx, accAddr)
		if err != nil {
			return nil, err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, commission)
		if err != nil {
			return nil, err
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
		),
	)

	return commission, nil
}

// GetTotalRewards returns the total amount of fee distribution rewards held in the store
func (k Keeper) GetTotalRewards(ctx context.Context) (totalRewards sdk.DecCoins) {
	k.IterateValidatorOutstandingRewards(ctx,
		func(_ sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
			totalRewards = totalRewards.Add(rewards.Rewards...)
			return false
		},
	)

	return totalRewards
}

// FundCommunityPool allows an account to directly fund the community fund pool.
// The amount is first added to the distribution module account and then directly
// added to the pool. An error is returned if the amount cannot be sent to the
// module account.
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount); err != nil {
		return err
	}

	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...)
	return k.FeePool.Set(ctx, feePool)
}

func (k Keeper) WithdrawSingleShareRecordReward(ctx context.Context, recordID uint64) error {
	record, err := k.stakingKeeper.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return err
	}

	ownerAddr, err := k.authKeeper.AddressCodec().StringToBytes(record.Owner)
	if err != nil {
		return err
	}
	owner := sdk.AccAddress(ownerAddr)

	if k.bankKeeper.BlockedAddr(owner) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", owner.String())
	}

	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
	if err != nil {
		return err
	}

	validatorFound := true
	_, err = k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		if !goerrors.Is(err, stakingtypes.ErrNoValidatorFound) {
			return err
		}

		validatorFound = false
	}

	delegationFound := true
	_, err = k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
	if err != nil {
		if !goerrors.Is(err, stakingtypes.ErrNoDelegation) {
			return err
		}

		delegationFound = false
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if validatorFound && delegationFound {
		// withdraw rewards into reward module account and send it to reward owner
		cacheCtx, write := sdkCtx.CacheContext()
		_, err = k.WithdrawDelegationRewards(cacheCtx, record.GetModuleAddress(), valAddr)
		if err != nil {
			return err
		}
		write()
	}

	// apply changes when the module account has positive balance
	balances := k.bankKeeper.GetAllBalances(ctx, record.GetModuleAddress())
	if !balances.Empty() {
		err = k.bankKeeper.SendCoins(ctx, record.GetModuleAddress(), owner, balances)
		if err != nil {
			return err
		}

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeWithdrawTokenizeShareReward,
				sdk.NewAttribute(types.AttributeKeyWithdrawAddress, owner.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
			),
		)
	}
	return nil
}

// withdraw reward for owning TokenizeShareRecord
func (k Keeper) WithdrawTokenizeShareRecordReward(ctx sdk.Context, ownerAddr sdk.AccAddress, recordID uint64) (sdk.Coins, error) {
	record, err := k.stakingKeeper.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return nil, err
	}

	if record.Owner != ownerAddr.String() {
		return nil, types.ErrNotTokenizeShareRecordOwner
	}

	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
	if err != nil {
		return nil, err
	}

	_, err = k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	_, err = k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
	if err != nil {
		return nil, err
	}

	// withdraw rewards into reward module account and send it to reward owner
	_, err = k.WithdrawDelegationRewards(ctx, record.GetModuleAddress(), valAddr)
	if err != nil {
		return nil, err
	}

	// apply changes when the module account has positive balance
	rewards := k.bankKeeper.GetAllBalances(ctx, record.GetModuleAddress())
	if !rewards.Empty() {
		err = k.bankKeeper.SendCoins(ctx, record.GetModuleAddress(), ownerAddr, rewards)
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawTokenizeShareReward,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, ownerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, rewards.String()),
		),
	)

	return rewards, nil
}

// withdraw reward for all owning TokenizeShareRecord
func (k Keeper) WithdrawAllTokenizeShareRecordReward(ctx sdk.Context, ownerAddr sdk.AccAddress) (sdk.Coins, error) {
	totalRewards := sdk.Coins{}

	records := k.stakingKeeper.GetTokenizeShareRecordsByOwner(ctx, ownerAddr)

	for _, record := range records {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(record.Validator)
		if err != nil {
			return nil, err
		}

		_, err = k.stakingKeeper.Validator(ctx, valAddr)
		if err != nil && !goerrors.Is(err, stakingtypes.ErrNoValidatorFound) {
			return nil, err
		}

		// if the error is ErrNoValidatorFound
		if err != nil {
			continue
		}

		_, err = k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
		if err != nil && !goerrors.Is(err, stakingtypes.ErrNoDelegation) {
			return nil, err
		}

		// if the error is ErrNoDelegation
		if err != nil {
			continue
		}

		// withdraw rewards into reward module account and send it to reward owner
		cacheCtx, write := ctx.CacheContext()
		_, err = k.WithdrawDelegationRewards(cacheCtx, record.GetModuleAddress(), valAddr)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			continue
		}

		// apply changes when the module account has positive balance
		balances := k.bankKeeper.GetAllBalances(cacheCtx, record.GetModuleAddress())
		if !balances.Empty() {
			err = k.bankKeeper.SendCoins(cacheCtx, record.GetModuleAddress(), ownerAddr, balances)
			if err != nil {
				k.Logger(ctx).Error(err.Error())
				continue
			}
			write()
			totalRewards = totalRewards.Add(balances...)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawTokenizeShareReward,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, ownerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, totalRewards.String()),
		),
	)

	return totalRewards, nil
}
