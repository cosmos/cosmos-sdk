package keeper

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
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
	// DelegatorsWithdrawAddress key: delAddr | value: withdrawAddr
	DelegatorsWithdrawAddress collections.Map[sdk.AccAddress, sdk.AccAddress]
	// ValidatorCurrentRewards key: valAddr | value: ValidatorCurrentRewards
	ValidatorCurrentRewards collections.Map[sdk.ValAddress, types.ValidatorCurrentRewards]
	// DelegatorStartingInfo key: valAddr+delAccAddr | value: DelegatorStartingInfo
	DelegatorStartingInfo collections.Map[collections.Pair[sdk.ValAddress, sdk.AccAddress], types.DelegatorStartingInfo]
	// ValidatorsAccumulatedCommission key: valAddr | value: ValidatorAccumulatedCommission
	ValidatorsAccumulatedCommission collections.Map[sdk.ValAddress, types.ValidatorAccumulatedCommission]
	// ValidatorOutstandingRewards key: valAddr | value: ValidatorOustandingRewards
	ValidatorOutstandingRewards collections.Map[sdk.ValAddress, types.ValidatorOutstandingRewards]
	// ValidatorHistoricalRewards key: valAddr+period | value: ValidatorHistoricalRewards
	ValidatorHistoricalRewards collections.Map[collections.Pair[sdk.ValAddress, uint64], types.ValidatorHistoricalRewards]
	PreviousProposer           collections.Item[sdk.ConsAddress]
	// ValidatorSlashEvents key: valAddr+height+period | value: ValidatorSlashEvent
	ValidatorSlashEvents collections.Map[collections.Triple[sdk.ValAddress, uint64, uint64], types.ValidatorSlashEvent]

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
		DelegatorsWithdrawAddress: collections.NewMap(
			sb,
			types.DelegatorWithdrawAddrPrefix,
			"delegators_withdraw_address",
			sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collcodec.KeyToValueCodec(sdk.AccAddressKey),
		),
		ValidatorCurrentRewards: collections.NewMap(
			sb,
			types.ValidatorCurrentRewardsPrefix,
			"validators_current_rewards",
			sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.ValidatorCurrentRewards](cdc),
		),
		DelegatorStartingInfo: collections.NewMap(
			sb,
			types.DelegatorStartingInfoPrefix,
			"delegators_starting_info",
			collections.PairKeyCodec(sdk.ValAddressKey, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.DelegatorStartingInfo](cdc),
		),
		ValidatorsAccumulatedCommission: collections.NewMap(
			sb,
			types.ValidatorAccumulatedCommissionPrefix,
			"validators_accumulated_commission",
			sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.ValidatorAccumulatedCommission](cdc),
		),
		ValidatorOutstandingRewards: collections.NewMap(
			sb,
			types.ValidatorOutstandingRewardsPrefix,
			"validator_outstanding_rewards",
			sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.ValidatorOutstandingRewards](cdc),
		),

		ValidatorHistoricalRewards: collections.NewMap(
			sb,
			types.ValidatorHistoricalRewardsPrefix,
			"validator_historical_rewards",
			collections.PairKeyCodec(sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), sdk.LEUint64Key), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.ValidatorHistoricalRewards](cdc),
		),
		PreviousProposer: collections.NewItem(sb, types.ProposerKey, "previous_proposer", collcodec.KeyToValueCodec(sdk.ConsAddressKey)),
		ValidatorSlashEvents: collections.NewMap(
			sb,
			types.ValidatorSlashEventPrefix,
			"validator_slash_events",
			collections.TripleKeyCodec(sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), collections.Uint64Key, collections.Uint64Key), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.ValidatorSlashEvent](cdc),
		),
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

	return k.DelegatorsWithdrawAddress.Set(ctx, delegatorAddr, withdrawAddr)
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
	accumCommission, err := k.ValidatorsAccumulatedCommission.Get(ctx, valAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}

	if accumCommission.Commission.IsZero() {
		return nil, types.ErrNoValidatorCommission
	}

	commission, remainder := accumCommission.Commission.TruncateDecimal()
	err = k.ValidatorsAccumulatedCommission.Set(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: remainder}) // leave remainder to withdraw later
	if err != nil {
		return nil, err
	}
	// update outstanding
	outstanding, err := k.ValidatorOutstandingRewards.Get(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	err = k.ValidatorOutstandingRewards.Set(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: outstanding.Rewards.Sub(sdk.NewDecCoinsFromCoins(commission...))})
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
	err := k.ValidatorOutstandingRewards.Walk(ctx, nil, func(_ sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool, err error) {
		totalRewards = totalRewards.Add(rewards.Rewards...)
		return false, nil
	},
	)
	if err != nil {
		panic(err)
	}

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
