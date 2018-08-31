package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// return a specific delegation
func (k Keeper) GetDelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress) (
	delegation types.Delegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	key := GetDelegationKey(delAddr, valAddr)
	value := store.Get(key)
	if value == nil {
		return delegation, false
	}

	delegation = types.MustUnmarshalDelegation(k.cdc, key, value)
	return delegation, true
}

// return all delegations used during genesis dump
func (k Keeper) GetAllDelegations(ctx sdk.Context) (delegations []types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegationKey)

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Key(), iterator.Value())
		delegations = append(delegations, delegation)
	}
	iterator.Close()
	return delegations
}

// Return all validators that a delegator is bonded to. If maxRetrieve is supplied, the respective amount will be returned
func (k Keeper) GetDelegatorBechValidators(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	maxRetrieve ...int16) (validators []types.BechValidator) {

	retrieve := len(maxRetrieve) > 0
	if retrieve {
		validators = make([]types.BechValidator, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegatorAddr)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		addr := iterator.Key()
		delegation := types.MustUnmarshalDelegation(k.cdc, addr, iterator.Value())
		validator, found := k.GetValidator(ctx, delegation.ValidatorAddr)
		if !found {
			panic(types.ErrNoValidatorFound(types.DefaultCodespace))
		}

		bechValidator, err := validator.Bech32Validator()
		if err != nil {
			panic(err.Error())
		}

		if retrieve {
			validators[i] = bechValidator
		} else {
			validators = append(validators, bechValidator)
		}
		i++
	}
	iterator.Close()
	return validators[:i] // trim
}

// Return all validators that a delegator is bonded to. If maxRetrieve is supplied, the respective amount will be returned
func (k Keeper) GetDelegatorValidators(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	maxRetrieve ...int16) (validators []types.Validator) {

	retrieve := len(maxRetrieve) > 0
	if retrieve {
		validators = make([]types.Validator, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegatorAddr)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		addr := iterator.Key()
		delegation := types.MustUnmarshalDelegation(k.cdc, addr, iterator.Value())
		validator, found := k.GetValidator(ctx, delegation.ValidatorAddr)
		if !found {
			panic(types.ErrNoValidatorFound(types.DefaultCodespace))
		}

		if retrieve {
			validators[i] = validator
		} else {
			validators = append(validators, validator)
		}

		i++
	}
	iterator.Close()
	return validators[:i] // trim
}

// return a validator that a delegator is bonded to
func (k Keeper) GetDelegatorBechValidator(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress) (bechValidator types.BechValidator) {

	validator := k.GetDelegatorValidator(ctx, delegatorAddr, validatorAddr)
	bechValidator, err := validator.Bech32Validator()
	if err != nil {
		panic(err.Error())
	}

	return
}

// return a validator that a delegator is bonded to
func (k Keeper) GetDelegatorValidator(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress) (validator types.Validator) {

	delegation, found := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		panic(types.ErrNoDelegation(types.DefaultCodespace))
	}

	validator, found = k.GetValidator(ctx, delegation.ValidatorAddr)
	if !found {
		panic(types.ErrNoValidatorFound(types.DefaultCodespace))
	}
	return
}

// return all delegations for a delegator. If maxRetrieve is supplied, the respective amount will be returned
func (k Keeper) GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve ...int16) (delegations []types.Delegation) {
	retrieve := len(maxRetrieve) > 0
	if retrieve {
		delegations = make([]types.Delegation, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Key(), iterator.Value())
		if retrieve {
			delegations[i] = delegation
		} else {
			delegations = append(delegations, delegation)
		}
		i++
	}
	iterator.Close()
	return delegations[:i] // trim
}

// Return all delegations for a delegator. If maxRetrieve is supplied, the respective amount will be returned
func (k Keeper) GetDelegatorDelegationsWithoutRat(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve ...int16) (delegations []types.DelegationWithoutDec) {
	retrieve := len(maxRetrieve) > 0
	if retrieve {
		delegations = make([]types.DelegationWithoutDec, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Key(), iterator.Value())
		delegationWithoutRat := types.NewDelegationWithoutDec(delegation)
		if retrieve {
			delegations[i] = delegationWithoutRat
		} else {
			delegations = append(delegations, delegationWithoutRat)
		}
		i++
	}
	iterator.Close()
	return delegations[:i] // trim
}

// set the delegation
func (k Keeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	b := types.MustMarshalDelegation(k.cdc, delegation)
	store.Set(GetDelegationKey(delegation.DelegatorAddr, delegation.ValidatorAddr), b)
}

// remove a delegation from store
func (k Keeper) RemoveDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegationKey(delegation.DelegatorAddr, delegation.ValidatorAddr))
}

//_____________________________________________________________________________________

// Return all unbonding delegations for a delegator. If maxRetrieve is supplied, the respective amount will be returned
func (k Keeper) GetDelegatorUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve ...int16) (unbondingDelegations []types.UnbondingDelegation) {

	retrieve := len(maxRetrieve) > 0
	if retrieve {
		unbondingDelegations = make([]types.UnbondingDelegation, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetUBDsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		unbondingDelegation := types.MustUnmarshalUBD(k.cdc, iterator.Key(), iterator.Value())
		if retrieve {
			unbondingDelegations[i] = unbondingDelegation
		} else {
			unbondingDelegations = append(unbondingDelegations, unbondingDelegation)
		}
		i++
	}
	iterator.Close()
	return unbondingDelegations[:i] // trim
}

// return a unbonding delegation
func (k Keeper) GetUnbondingDelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress) (ubd types.UnbondingDelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	key := GetUBDKey(delAddr, valAddr)
	value := store.Get(key)
	if value == nil {
		return ubd, false
	}

	ubd = types.MustUnmarshalUBD(k.cdc, key, value)
	return ubd, true
}

// return all unbonding delegations from a particular validator
func (k Keeper) GetUnbondingDelegationsFromValidator(ctx sdk.Context, valAddr sdk.ValAddress) (ubds []types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, GetUBDsByValIndexKey(valAddr))

	for ; iterator.Valid(); iterator.Next() {
		key := GetUBDKeyFromValIndexKey(iterator.Key())
		value := store.Get(key)
		ubd := types.MustUnmarshalUBD(k.cdc, key, value)
		ubds = append(ubds, ubd)
	}

	iterator.Close()
	return ubds
}

// iterate through all of the unbonding delegations
func (k Keeper) IterateUnbondingDelegations(ctx sdk.Context, fn func(index int64, ubd types.UnbondingDelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, UnbondingDelegationKey)

	for i := int64(0); iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(k.cdc, iterator.Key(), iterator.Value())
		stop := fn(i, ubd)
		if stop {
			break
		}
		i++
	}
	iterator.Close()
}

// set the unbonding delegation and associated index
func (k Keeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalUBD(k.cdc, ubd)
	key := GetUBDKey(ubd.DelegatorAddr, ubd.ValidatorAddr)
	store.Set(key, bz)
	store.Set(GetUBDByValIndexKey(ubd.DelegatorAddr, ubd.ValidatorAddr), []byte{}) // index, store empty bytes
}

// remove the unbonding delegation object and associated index
func (k Keeper) RemoveUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	key := GetUBDKey(ubd.DelegatorAddr, ubd.ValidatorAddr)
	store.Delete(key)
	store.Delete(GetUBDByValIndexKey(ubd.DelegatorAddr, ubd.ValidatorAddr))
}

//_____________________________________________________________________________________

// Return all redelegations for a delegator. If maxRetrieve is supplied, the respective amount will be returned
func (k Keeper) GetRedelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve ...int16) (redelegations []types.Redelegation) {

	retrieve := len(maxRetrieve) > 0
	if retrieve {
		redelegations = make([]types.Redelegation, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetREDsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		redelegation := types.MustUnmarshalRED(k.cdc, iterator.Key(), iterator.Value())
		if retrieve {
			redelegations[i] = redelegation
		} else {
			redelegations = append(redelegations, redelegation)
		}
		i++
	}
	iterator.Close()
	return redelegations[:i] // trim
}

// return a redelegation
func (k Keeper) GetRedelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) (red types.Redelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	key := GetREDKey(delAddr, valSrcAddr, valDstAddr)
	value := store.Get(key)
	if value == nil {
		return red, false
	}

	red = types.MustUnmarshalRED(k.cdc, key, value)
	return red, true
}

// return all redelegations from a particular validator
func (k Keeper) GetRedelegationsFromValidator(ctx sdk.Context, valAddr sdk.ValAddress) (reds []types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, GetREDsFromValSrcIndexKey(valAddr))
	for ; iterator.Valid(); iterator.Next() {
		key := GetREDKeyFromValSrcIndexKey(iterator.Key())
		value := store.Get(key)
		red := types.MustUnmarshalRED(k.cdc, key, value)
		reds = append(reds, red)
	}
	iterator.Close()
	return reds
}

// check if validator has an incoming redelegation
func (k Keeper) HasReceivingRedelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool {

	store := ctx.KVStore(k.storeKey)
	prefix := GetREDsByDelToValDstIndexKey(delAddr, valDstAddr)
	iterator := sdk.KVStorePrefixIterator(store, prefix) //smallest to largest

	found := false
	if iterator.Valid() {
		//record found
		found = true
	}
	iterator.Close()
	return found
}

// set a redelegation and associated index
func (k Keeper) SetRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalRED(k.cdc, red)
	key := GetREDKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr)
	store.Set(key, bz)
	store.Set(GetREDByValSrcIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr), []byte{})
	store.Set(GetREDByValDstIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr), []byte{})
}

// remove a redelegation object and associated index
func (k Keeper) RemoveRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	redKey := GetREDKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr)
	store.Delete(redKey)
	store.Delete(GetREDByValSrcIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr))
	store.Delete(GetREDByValDstIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr))
}

//_____________________________________________________________________________________

// Perform a delegation, set/update everything necessary within the store.
func (k Keeper) Delegate(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Coin,
	validator types.Validator, subtractAccount bool) (newShares sdk.Dec, err sdk.Error) {

	// Get or create the delegator delegation
	delegation, found := k.GetDelegation(ctx, delAddr, validator.Operator)
	if !found {
		delegation = types.Delegation{
			DelegatorAddr: delAddr,
			ValidatorAddr: validator.Operator,
			Shares:        sdk.ZeroDec(),
		}
	}

	if subtractAccount {
		// Account new shares, save
		_, _, err = k.coinKeeper.SubtractCoins(ctx, delegation.DelegatorAddr, sdk.Coins{bondAmt})
		if err != nil {
			return
		}
	}

	pool := k.GetPool(ctx)
	validator, pool, newShares = validator.AddTokensFromDel(pool, bondAmt.Amount)
	delegation.Shares = delegation.Shares.Add(newShares)

	// Update delegation height
	delegation.Height = ctx.BlockHeight()

	k.SetPool(ctx, pool)
	k.SetDelegation(ctx, delegation)
	k.UpdateValidator(ctx, validator)

	return
}

// unbond the the delegation return
func (k Keeper) unbond(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	shares sdk.Dec) (amount sdk.Dec, err sdk.Error) {

	// check if delegation has any shares in it unbond
	delegation, found := k.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		err = types.ErrNoDelegatorForAddress(k.Codespace())
		return
	}

	// retrieve the amount to remove
	if delegation.Shares.LT(shares) {
		err = types.ErrNotEnoughDelegationShares(k.Codespace(), delegation.Shares.String())
		return
	}

	// get validator
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		err = types.ErrNoValidatorFound(k.Codespace())
		return
	}

	// subtract shares from delegator
	delegation.Shares = delegation.Shares.Sub(shares)

	// remove the delegation
	if delegation.Shares.IsZero() {

		// if the delegation is the operator of the validator then
		// trigger a jail validator
		if bytes.Equal(delegation.DelegatorAddr, validator.Operator) && validator.Jailed == false {
			validator.Jailed = true
		}
		k.RemoveDelegation(ctx, delegation)
	} else {
		// Update height
		delegation.Height = ctx.BlockHeight()
		k.SetDelegation(ctx, delegation)
	}

	// remove the coins from the validator
	pool := k.GetPool(ctx)
	validator, pool, amount = validator.RemoveDelShares(pool, shares)

	k.SetPool(ctx, pool)

	// update then remove validator if necessary
	validator = k.UpdateValidator(ctx, validator)
	if validator.DelegatorShares.IsZero() {
		k.RemoveValidator(ctx, validator.Operator)
	}

	return
}

//______________________________________________________________________________________________________

// complete unbonding an unbonding record
func (k Keeper) BeginUnbonding(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) sdk.Error {

	// TODO quick fix, instead we should use an index, see https://github.com/cosmos/cosmos-sdk/issues/1402
	_, found := k.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if found {
		return types.ErrExistingUnbondingDelegation(k.Codespace())
	}

	returnAmount, err := k.unbond(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return err
	}

	// create the unbonding delegation
	params := k.GetParams(ctx)
	minTime := ctx.BlockHeader().Time.Add(params.UnbondingTime)
	balance := sdk.NewCoin(params.BondDenom, returnAmount.RoundInt())

	ubd := types.UnbondingDelegation{
		DelegatorAddr:  delAddr,
		ValidatorAddr:  valAddr,
		MinTime:        minTime,
		Balance:        balance,
		InitialBalance: balance,
	}
	k.SetUnbondingDelegation(ctx, ubd)
	return nil
}

// complete unbonding an unbonding record
func (k Keeper) CompleteUnbonding(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) sdk.Error {

	ubd, found := k.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if !found {
		return types.ErrNoUnbondingDelegation(k.Codespace())
	}

	// ensure that enough time has passed
	ctxTime := ctx.BlockHeader().Time
	if ubd.MinTime.After(ctxTime) {
		return types.ErrNotMature(k.Codespace(), "unbonding", "unit-time", ubd.MinTime, ctxTime)
	}

	_, _, err := k.coinKeeper.AddCoins(ctx, ubd.DelegatorAddr, sdk.Coins{ubd.Balance})
	if err != nil {
		return err
	}
	k.RemoveUnbondingDelegation(ctx, ubd)
	return nil
}

// complete unbonding an unbonding record
func (k Keeper) BeginRedelegation(ctx sdk.Context, delAddr sdk.AccAddress,
	valSrcAddr, valDstAddr sdk.ValAddress, sharesAmount sdk.Dec) sdk.Error {

	// check if this is a transitive redelegation
	if k.HasReceivingRedelegation(ctx, delAddr, valSrcAddr) {
		return types.ErrTransitiveRedelegation(k.Codespace())
	}

	returnAmount, err := k.unbond(ctx, delAddr, valSrcAddr, sharesAmount)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)
	returnCoin := sdk.NewCoin(params.BondDenom, returnAmount.RoundInt())
	dstValidator, found := k.GetValidator(ctx, valDstAddr)
	if !found {
		return types.ErrBadRedelegationDst(k.Codespace())
	}
	sharesCreated, err := k.Delegate(ctx, delAddr, returnCoin, dstValidator, false)
	if err != nil {
		return err
	}

	// create the unbonding delegation
	minTime := ctx.BlockHeader().Time.Add(params.UnbondingTime)

	red := types.Redelegation{
		DelegatorAddr:    delAddr,
		ValidatorSrcAddr: valSrcAddr,
		ValidatorDstAddr: valDstAddr,
		MinTime:          minTime,
		SharesDst:        sharesCreated,
		SharesSrc:        sharesAmount,
		Balance:          returnCoin,
		InitialBalance:   returnCoin,
	}
	k.SetRedelegation(ctx, red)
	return nil
}

// complete unbonding an ongoing redelegation
func (k Keeper) CompleteRedelegation(ctx sdk.Context, delAddr sdk.AccAddress,
	valSrcAddr, valDstAddr sdk.ValAddress) sdk.Error {

	red, found := k.GetRedelegation(ctx, delAddr, valSrcAddr, valDstAddr)
	if !found {
		return types.ErrNoRedelegation(k.Codespace())
	}

	// ensure that enough time has passed
	ctxTime := ctx.BlockHeader().Time
	if red.MinTime.After(ctxTime) {
		return types.ErrNotMature(k.Codespace(), "redelegation", "unit-time", red.MinTime, ctxTime)
	}

	k.RemoveRedelegation(ctx, red)
	return nil
}
