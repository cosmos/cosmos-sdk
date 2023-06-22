package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper of the distribution store
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           codec.BinaryCodec
	paramSpace    paramtypes.Subspace
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper

	blockedAddrs map[string]bool

	feeCollectorName string // name of the FeeCollector ModuleAccount
}

// NewKeeper creates a new distribution Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper,
	feeCollectorName string, blockedAddrs map[string]bool,
) Keeper {
	// ensure distribution module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		paramSpace:       paramSpace,
		authKeeper:       ak,
		bankKeeper:       bk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollectorName,
		blockedAddrs:     blockedAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetWithdrawAddr sets a new address that will receive the rewards upon withdrawal
func (k Keeper) SetWithdrawAddr(ctx sdk.Context, delegatorAddr sdk.AccAddress, withdrawAddr sdk.AccAddress) error {
	if k.blockedAddrs[withdrawAddr.String()] {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive external funds", withdrawAddr)
	}

	if !k.GetWithdrawAddrEnabled(ctx) {
		return types.ErrSetWithdrawAddrDisabled
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetWithdrawAddress,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, withdrawAddr.String()),
		),
	)

	k.SetDelegatorWithdrawAddr(ctx, delegatorAddr, withdrawAddr)
	return nil
}

// withdraw rewards from a delegation
func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	val := k.stakingKeeper.Validator(ctx, valAddr)
	if val == nil {
		return nil, types.ErrNoValidatorDistInfo
	}

	del := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	if del == nil {
		return nil, types.ErrEmptyDelegationDistInfo
	}

	// withdraw rewards
	rewards, err := k.withdrawDelegationRewards(ctx, val, del)
	if err != nil {
		return nil, err
	}

	// reinitialize the delegation
	k.initializeDelegation(ctx, valAddr, delAddr)
	return rewards, nil
}

// withdraw validator commission
func (k Keeper) WithdrawValidatorCommission(ctx sdk.Context, valAddr sdk.ValAddress) (sdk.Coins, error) {
	// fetch validator accumulated commission
	accumCommission := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if accumCommission.Commission.IsZero() {
		return nil, types.ErrNoValidatorCommission
	}

	commission, remainder := accumCommission.Commission.TruncateDecimal()
	k.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: remainder}) // leave remainder to withdraw later

	// update outstanding
	outstanding := k.GetValidatorOutstandingRewards(ctx, valAddr).Rewards
	k.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: outstanding.Sub(sdk.NewDecCoinsFromCoins(commission...))})

	if !commission.IsZero() {
		accAddr := sdk.AccAddress(valAddr)
		withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, commission)
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
		),
	)

	return commission, nil
}

// GetTotalRewards returns the total amount of fee distribution rewards held in the store
func (k Keeper) GetTotalRewards(ctx sdk.Context) (totalRewards sdk.DecCoins) {
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
func (k Keeper) FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount); err != nil {
		return err
	}

	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...)
	k.SetFeePool(ctx, feePool)

	return nil
}

func (k Keeper) WithdrawSingleShareRecordReward(ctx sdk.Context, recordID uint64) error {
	record, err := k.stakingKeeper.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return err
	}

	owner, err := sdk.AccAddressFromBech32(record.Owner)
	if err != nil {
		return err
	}

	valAddr, err := sdk.ValAddressFromBech32(record.Validator)
	if err != nil {
		return err
	}

	val := k.stakingKeeper.Validator(ctx, valAddr)
	del := k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
	if val != nil && del != nil {
		// withdraw rewards into reward module account and send it to reward owner
		cacheCtx, write := ctx.CacheContext()
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

		ctx.EventManager().EmitEvent(
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

	valAddr, err := sdk.ValAddressFromBech32(record.Validator)
	if err != nil {
		return nil, err
	}

	val := k.stakingKeeper.Validator(ctx, valAddr)
	if val == nil {
		return nil, err
	}

	del := k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
	if del == nil {
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
		valAddr, err := sdk.ValAddressFromBech32(record.Validator)
		if err != nil {
			return nil, err
		}

		val := k.stakingKeeper.Validator(ctx, valAddr)
		if val == nil {
			continue
		}

		del := k.stakingKeeper.Delegation(ctx, record.GetModuleAddress(), valAddr)
		if del == nil {
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
