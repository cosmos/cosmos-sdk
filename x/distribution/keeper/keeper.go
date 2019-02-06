package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// keeper of the staking store
type Keeper struct {
	storeKey            sdk.StoreKey
	cdc                 *codec.Codec
	paramSpace          params.Subspace
	bankKeeper          types.BankKeeper
	stakingKeeper       types.StakingKeeper
	feeCollectionKeeper types.FeeCollectionKeeper

	// codespace
	codespace sdk.CodespaceType
}

// create a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace, ck types.BankKeeper,
	sk types.StakingKeeper, fck types.FeeCollectionKeeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:            key,
		cdc:                 cdc,
		paramSpace:          paramSpace.WithKeyTable(ParamKeyTable()),
		bankKeeper:          ck,
		stakingKeeper:       sk,
		feeCollectionKeeper: fck,
		codespace:           codespace,
	}
	return keeper
}

// set withdraw address
func (k Keeper) SetWithdrawAddr(ctx sdk.Context, delegatorAddr sdk.AccAddress, withdrawAddr sdk.AccAddress) sdk.Error {
	if !k.GetWithdrawAddrEnabled(ctx) {
		return types.ErrSetWithdrawAddrDisabled(k.codespace)
	}

	k.SetDelegatorWithdrawAddr(ctx, delegatorAddr, withdrawAddr)

	return nil
}

// withdraw rewards from a delegation
func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) sdk.Error {
	val := k.stakingKeeper.Validator(ctx, valAddr)
	if val == nil {
		return types.ErrNoValidatorDistInfo(k.codespace)
	}

	del := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	if del == nil {
		return types.ErrNoDelegationDistInfo(k.codespace)
	}

	// withdraw rewards
	if err := k.withdrawDelegationRewards(ctx, val, del); err != nil {
		return err
	}

	// reinitialize the delegation
	k.initializeDelegation(ctx, valAddr, delAddr)

	return nil
}

// withdraw validator commission
func (k Keeper) WithdrawValidatorCommission(ctx sdk.Context, valAddr sdk.ValAddress) sdk.Error {

	// fetch validator accumulated commission
	commission := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if commission.IsZero() {
		return types.ErrNoValidatorCommission(k.codespace)
	}

	coins, remainder := commission.TruncateDecimal()

	// leave remainder to withdraw later
	k.SetValidatorAccumulatedCommission(ctx, valAddr, remainder)

	// update outstanding
	outstanding := k.GetOutstandingRewards(ctx)
	k.SetOutstandingRewards(ctx, outstanding.Minus(sdk.NewDecCoins(coins)))

	accAddr := sdk.AccAddress(valAddr)
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)

	if _, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coins); err != nil {
		return err
	}

	return nil
}
