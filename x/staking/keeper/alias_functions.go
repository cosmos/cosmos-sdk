package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// Implements ValidatorSet
var _ sdk.ValidatorSet = Keeper{}

// iterate through the validator set and perform the provided function
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

// iterate through the bonded validator set and perform the provided function
func (k Keeper) IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	maxValidators := k.MaxValidators(ctx)

	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey)
	defer iterator.Close()

	i := int64(0)
	for ; iterator.Valid() && i < int64(maxValidators); iterator.Next() {
		address := iterator.Value()
		validator := k.mustGetValidator(ctx, address)

		if validator.Status == sdk.Bonded {
			stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
			if stop {
				break
			}
			i++
		}
	}
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateLastValidators(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	iterator := k.LastValidatorsIterator(ctx)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := AddressFromLastValidatorPowerKey(iterator.Key())
		validator, found := k.GetValidator(ctx, address)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", address))
		}

		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

// get the sdk.validator for a particular address
func (k Keeper) Validator(ctx sdk.Context, address sdk.ValAddress) sdk.Validator {
	val, found := k.GetValidator(ctx, address)
	if !found {
		return nil
	}
	return val
}

// get the sdk.validator for a particular pubkey
func (k Keeper) ValidatorByConsAddr(ctx sdk.Context, addr sdk.ConsAddress) sdk.Validator {
	val, found := k.GetValidatorByConsAddr(ctx, addr)
	if !found {
		return nil
	}
	return val
}

// TotalBondedTokens total staking tokens supply which is bonded
func (k Keeper) TotalBondedTokens(ctx sdk.Context) sdk.Int {
	params := k.GetParams(ctx)

	bondedPool, err := k.supplyKeeper.GetPoolAccountByName(ctx, BondedTokensName)
	if err != nil {
		panic(err)
	}

	return bondedPool.GetCoins().AmountOf(params.BondDenom)
}

// StakingTokenSupply total staking tokens supply bonded and unbonded
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	params := k.GetParams(ctx)
	unbondedPool, err := k.supplyKeeper.GetPoolAccountByName(ctx, UnbondedTokensName)
	if err != nil {
		panic(err)
	}

	bondedPool, err := k.supplyKeeper.GetPoolAccountByName(ctx, BondedTokensName)
	if err != nil {
		panic(err)
	}

	return bondedPool.GetCoins().AmountOf(params.BondDenom).Add(unbondedPool.GetCoins().AmountOf(params.BondDenom))
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	params := k.GetParams(ctx)
	bondedPool, err := k.supplyKeeper.GetPoolAccountByName(ctx, BondedTokensName)
	if err != nil {
		panic(err)
	}

	stakeSupply := k.StakingTokenSupply(ctx)
	if stakeSupply.IsPositive() {
		return bondedPool.GetCoins().AmountOf(params.BondDenom).ToDec().QuoInt(stakeSupply)
	}
	return sdk.ZeroDec()
}

// Implements DelegationSet

var _ sdk.DelegationSet = Keeper{}

// Returns self as it is both a validatorset and delegationset
func (k Keeper) GetValidatorSet() sdk.ValidatorSet {
	return k
}

// get the delegation for a particular set of delegator and validator addresses
func (k Keeper) Delegation(ctx sdk.Context, addrDel sdk.AccAddress, addrVal sdk.ValAddress) sdk.Delegation {
	bond, ok := k.GetDelegation(ctx, addrDel, addrVal)
	if !ok {
		return nil
	}

	return bond
}

// iterate through all of the delegations from a delegator
func (k Keeper) IterateDelegations(ctx sdk.Context, delAddr sdk.AccAddress,
	fn func(index int64, del sdk.Delegation) (stop bool)) {

	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delAddr)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) // smallest to largest
	defer iterator.Close()
	for i := int64(0); iterator.Valid(); iterator.Next() {
		del := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		stop := fn(i, del)
		if stop {
			break
		}
		i++
	}
}

// return all delegations used during genesis dump
// TODO: remove this func, change all usage for iterate functionality
func (k Keeper) GetAllSDKDelegations(ctx sdk.Context) (delegations []sdk.Delegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		delegations = append(delegations, delegation)
	}
	return delegations
}

// GetPools returns the bonded and unbonded tokens pool accounts
func (k Keeper) GetPools(ctx sdk.Context) (bondedPool, unbondedPool supply.PoolAccount) {
	bondedPool, err := k.supplyKeeper.GetPoolAccountByName(ctx, BondedTokensName)
	if err != nil {
		return
	}

	unbondedPool, err = k.supplyKeeper.GetPoolAccountByName(ctx, BondedTokensName)
	if err != nil {
		return
	}

	return bondedPool, unbondedPool
}

// SetBondedPool sets the bonded tokens pool account
func (k Keeper) SetBondedPool(ctx sdk.Context, newBondPool supply.PoolAccount) {
	if newBondPool.Name() != BondedTokensName {
		panic(fmt.Sprintf("invalid name for bonded pool (%s ≠ %s)", BondedTokensName, newBondPool.Name()))
	}
	k.supplyKeeper.SetPoolAccount(ctx, newBondPool)
}

// SetUnbondedPool sets the unbonded tokens pool account
func (k Keeper) SetUnbondedPool(ctx sdk.Context, newUnbondPool supply.PoolAccount) {
	if newUnbondPool.Name() != UnbondedTokensName {
		panic(fmt.Sprintf("invalid name for unbonded pool (%s ≠ %s)", UnbondedTokensName, newUnbondPool.Name()))
	}
	k.supplyKeeper.SetPoolAccount(ctx, newUnbondPool)
}
