package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
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

	externalCommunityPool types.ExternalCommunityPoolKeeper
}

type InitOption func(*Keeper)

// WithExternalCommunityPool will enable the external pool functionality in x/distribution, directing
// community pool funds to the provided keeper.
//
// WARNING: using an external community pool will cause the following handlers to error when called:
// - FundCommunityPool tx
// - CommunityPoolSpend tx
// - CommunityPool query
func WithExternalCommunityPool(poolKeeper types.ExternalCommunityPoolKeeper) InitOption {
	return func(k *Keeper) {
		k.externalCommunityPool = poolKeeper
	}
}

// NewKeeper creates a new distribution Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	feeCollectorName, authority string,
	opts ...InitOption,
) Keeper {
	// ensure distribution module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		storeService:          storeService,
		cdc:                   cdc,
		authKeeper:            ak,
		bankKeeper:            bk,
		stakingKeeper:         sk,
		feeCollectorName:      feeCollectorName,
		authority:             authority,
		Params:                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		FeePool:               collections.NewItem(sb, types.FeePoolKey, "fee_pool", codec.CollValue[types.FeePool](cdc)),
		externalCommunityPool: nil,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	for _, opt := range opts {
		opt(&k)
	}

	if k.HasExternalCommunityPool() {
		// ensure external module account is set if we are enabling it
		// this will ensure that funds can be transferred to it.
		if addr := ak.GetModuleAddress(k.externalCommunityPool.GetCommunityPoolModule()); addr == nil {
			panic(fmt.Sprintf("%s module account has not been set", k.externalCommunityPool.GetCommunityPoolModule()))
		}
	}

	return k
}

// GetAuthority returns the x/distribution module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// HasExternalCommunityPool is a helper function to denote whether the x/distribution module
// is using its native community pool, or using an external pool.
func (k Keeper) HasExternalCommunityPool() bool {
	return k.externalCommunityPool != nil
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

	return k.SetDelegatorWithdrawAddr(ctx, delegatorAddr, withdrawAddr)
}

// WithdrawDelegationRewards withdraws rewards from a delegation
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

// WithdrawValidatorCommission withdraws validator commission.
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
	err = k.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: remainder}) // leave remainder to withdraw later
	if err != nil {
		return nil, err
	}

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
