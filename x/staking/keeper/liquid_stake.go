package keeper

import (
	"context"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SetTotalLiquidStakedTokens stores the total outstanding tokens owned by a liquid staking provider
func (k Keeper) SetTotalLiquidStakedTokens(ctx context.Context, tokens math.Int) {
	store := k.storeService.OpenKVStore(ctx)

	tokensBz, err := tokens.Marshal()
	if err != nil {
		panic(err)
	}

	err = store.Set(types.TotalLiquidStakedTokensKey, tokensBz)
	if err != nil {
		panic(err)
	}
}

// GetTotalLiquidStakedTokens returns the total outstanding tokens owned by a liquid staking provider
// Returns zero if the total liquid stake amount has not been initialized
func (k Keeper) GetTotalLiquidStakedTokens(ctx context.Context) math.Int {
	store := k.storeService.OpenKVStore(ctx)
	tokensBz, err := store.Get(types.TotalLiquidStakedTokensKey)
	if err != nil {
		panic(err)
	}

	if tokensBz == nil {
		return math.ZeroInt()
	}

	var tokens math.Int
	if err := tokens.Unmarshal(tokensBz); err != nil {
		panic(err)
	}

	return tokens
}

// Checks if an account associated with a given delegation is related to liquid staking
//
// This is determined by checking if the account has a 32-length address
// which will identify the following scenarios:
//   - An account has tokenized their shares, and thus the delegation is
//     owned by the tokenize share record module account
//   - A liquid staking provider is delegating through an ICA account
//
// Both ICA accounts and tokenize share record module accounts have 32-length addresses
// NOTE: This will have to be refactored before adapting it to chains beyond gaia
// as other chains may have 32-length addresses that are not related to the above scenarios
func (k Keeper) DelegatorIsLiquidStaker(delegatorAddress sdk.AccAddress) bool {
	return len(delegatorAddress) == 32
}

// CheckExceedsGlobalLiquidStakingCap checks if a liquid delegation would cause the
// global liquid staking cap to be exceeded
// A liquid delegation is defined as either tokenized shares, or a delegation from an ICA Account
// The total stake is determined by the balance of the bonded pool
// If the delegation's shares are already bonded (e.g. in the event of a tokenized share)
// the tokens are already included in the bonded pool
// If the delegation's shares are not bonded (e.g. normal delegation),
// we need to add the tokens to the current bonded pool balance to get the total staked
func (k Keeper) CheckExceedsGlobalLiquidStakingCap(ctx context.Context, tokens math.Int, sharesAlreadyBonded bool) (bool, error) {
	liquidStakingCap, err := k.GlobalLiquidStakingCap(ctx)
	if err != nil {
		return false, err
	}

	liquidStakedAmount := k.GetTotalLiquidStakedTokens(ctx)

	// Determine the total stake from the balance of the bonded pool
	// If this is not a tokenized delegation, we need to add the tokens to the pool balance since
	// they would not have been counted yet
	// If this is for a tokenized delegation, the tokens are already included in the pool balance
	totalStakedAmount, err := k.TotalBondedTokens(ctx)
	if err != nil {
		return false, err
	}

	if !sharesAlreadyBonded {
		totalStakedAmount = totalStakedAmount.Add(tokens)
	}

	// Calculate the percentage of stake that is liquid

	updatedLiquidStaked := math.LegacyNewDecFromInt(liquidStakedAmount.Add(tokens))
	liquidStakePercent := updatedLiquidStaked.Quo(math.LegacyNewDecFromInt(totalStakedAmount))

	return liquidStakePercent.GT(liquidStakingCap), nil
}

// CheckExceedsValidatorBondCap checks if a liquid delegation to a validator would cause
// the liquid shares to exceed the validator bond factor
// A liquid delegation is defined as either tokenized shares, or a delegation from an ICA Account
// Returns true if the cap is exceeded
func (k Keeper) CheckExceedsValidatorBondCap(ctx context.Context, validator types.Validator, shares math.LegacyDec) (bool, error) {
	validatorBondFactor, err := k.ValidatorBondFactor(ctx)
	if err != nil {
		return false, err
	}

	if validatorBondFactor.Equal(types.ValidatorBondCapDisabled) {
		return false, nil
	}
	maxValLiquidShares := validator.ValidatorBondShares.Mul(validatorBondFactor)

	return validator.LiquidShares.Add(shares).GT(maxValLiquidShares), nil
}

// CheckExceedsValidatorLiquidStakingCap checks if a liquid delegation could cause the
// total liquid shares to exceed the liquid staking cap
// A liquid delegation is defined as either tokenized shares, or a delegation from an ICA Account
// If the liquid delegation's shares are already bonded (e.g. in the event of a tokenized share)
// the tokens are already included in the validator's delegator shares
// If the liquid delegation's shares are not bonded (e.g. normal delegation),
// we need to add the shares to the current validator's delegator shares to get the total shares
// Returns true if the cap is exceeded
func (k Keeper) CheckExceedsValidatorLiquidStakingCap(ctx context.Context, validator types.Validator, shares math.LegacyDec, sharesAlreadyBonded bool) (bool, error) {
	updatedLiquidShares := validator.LiquidShares.Add(shares)

	updatedTotalShares := validator.DelegatorShares
	if !sharesAlreadyBonded {
		updatedTotalShares = updatedTotalShares.Add(shares)
	}

	liquidStakePercent := updatedLiquidShares.Quo(updatedTotalShares)
	liquidStakingCap, err := k.ValidatorLiquidStakingCap(ctx)
	if err != nil {
		return false, err
	}

	return liquidStakePercent.GT(liquidStakingCap), nil
}

// SafelyIncreaseTotalLiquidStakedTokens increments the total liquid staked tokens
// if the global cap is not surpassed by this delegation
//
// The percentage of liquid staked tokens must be less than the GlobalLiquidStakingCap:
// (TotalLiquidStakedTokens / TotalStakedTokens) <= GlobalLiquidStakingCap
func (k Keeper) SafelyIncreaseTotalLiquidStakedTokens(ctx context.Context, amount math.Int, sharesAlreadyBonded bool) error {
	exceedsCap, err := k.CheckExceedsGlobalLiquidStakingCap(ctx, amount, sharesAlreadyBonded)
	if err != nil {
		return err
	}

	if exceedsCap {
		return types.ErrGlobalLiquidStakingCapExceeded
	}

	k.SetTotalLiquidStakedTokens(ctx, k.GetTotalLiquidStakedTokens(ctx).Add(amount))
	return nil
}

// DecreaseTotalLiquidStakedTokens decrements the total liquid staked tokens
func (k Keeper) DecreaseTotalLiquidStakedTokens(ctx context.Context, amount math.Int) error {
	totalLiquidStake := k.GetTotalLiquidStakedTokens(ctx)
	if amount.GT(totalLiquidStake) {
		return types.ErrTotalLiquidStakedUnderflow
	}
	k.SetTotalLiquidStakedTokens(ctx, totalLiquidStake.Sub(amount))
	return nil
}

// SafelyIncreaseValidatorLiquidShares increments the liquid shares on a validator, if:
// the validator bond factor and validator liquid staking cap will not be exceeded by this delegation
//
// The percentage of validator liquid shares must be less than the ValidatorLiquidStakingCap,
// and the total liquid staked shares cannot exceed the validator bond cap
//  1. (TotalLiquidStakedTokens / TotalStakedTokens) <= ValidatorLiquidStakingCap
//  2. LiquidShares <= (ValidatorBondShares * ValidatorBondFactor)
func (k Keeper) SafelyIncreaseValidatorLiquidShares(ctx context.Context, valAddress sdk.ValAddress, shares math.LegacyDec, sharesAlreadyBonded bool) (types.Validator, error) {
	validator, err := k.GetValidator(ctx, valAddress)
	if err != nil {
		return validator, err
	}

	// Confirm the validator bond factor and validator liquid staking cap will not be exceeded
	exceedsValidatorBondCap, err := k.CheckExceedsValidatorBondCap(ctx, validator, shares)
	if err != nil {
		return validator, err
	}

	if exceedsValidatorBondCap {
		return validator, types.ErrInsufficientValidatorBondShares
	}

	exceedsValidatorLiquidStakingCap, err := k.CheckExceedsValidatorLiquidStakingCap(ctx, validator, shares, sharesAlreadyBonded)
	if err != nil {
		return validator, err
	}

	if exceedsValidatorLiquidStakingCap {
		return validator, types.ErrValidatorLiquidStakingCapExceeded
	}

	// Increment the validator's liquid shares
	validator.LiquidShares = validator.LiquidShares.Add(shares)
	err = k.SetValidator(ctx, validator)
	if err != nil {
		return types.Validator{}, err
	}

	return validator, nil
}

// DecreaseValidatorLiquidShares decrements the liquid shares on a validator
func (k Keeper) DecreaseValidatorLiquidShares(ctx context.Context, valAddress sdk.ValAddress, shares math.LegacyDec) (types.Validator, error) {
	validator, err := k.GetValidator(ctx, valAddress)
	if err != nil {
		return validator, err
	}

	if shares.GT(validator.LiquidShares) {
		return validator, types.ErrValidatorLiquidSharesUnderflow
	}

	validator.LiquidShares = validator.LiquidShares.Sub(shares)
	err = k.SetValidator(ctx, validator)
	if err != nil {
		return types.Validator{}, err
	}

	return validator, nil
}

// Increase validator bond shares increments the validator's self bond
// in the event that the delegation amount on a validator bond delegation is increased
func (k Keeper) IncreaseValidatorBondShares(ctx context.Context, valAddress sdk.ValAddress, shares math.LegacyDec) error {
	validator, err := k.GetValidator(ctx, valAddress)
	if err != nil {
		return err
	}

	validator.ValidatorBondShares = validator.ValidatorBondShares.Add(shares)
	err = k.SetValidator(ctx, validator)
	if err != nil {
		return err
	}

	return nil
}

// SafelyDecreaseValidatorBond decrements the validator's self bond
// so long as it will not cause the current delegations to exceed the threshold
// set by validator bond factor
func (k Keeper) SafelyDecreaseValidatorBond(ctx context.Context, valAddress sdk.ValAddress, shares math.LegacyDec) error {
	validator, err := k.GetValidator(ctx, valAddress)
	if err != nil {
		return err
	}

	// Check if the decreased self bond will cause the validator bond threshold to be exceeded
	validatorBondFactor, err := k.ValidatorBondFactor(ctx)
	if err != nil {
		return err
	}

	validatorBondEnabled := !validatorBondFactor.Equal(types.ValidatorBondCapDisabled)
	maxValTotalShare := validator.ValidatorBondShares.Sub(shares).Mul(validatorBondFactor)

	if validatorBondEnabled && validator.LiquidShares.GT(maxValTotalShare) {
		return types.ErrInsufficientValidatorBondShares
	}

	// Decrement the validator's self bond
	validator.ValidatorBondShares = validator.ValidatorBondShares.Sub(shares)
	err = k.SetValidator(ctx, validator)
	if err != nil {
		return err
	}

	return nil
}

// Adds a lock that prevents tokenizing shares for an account
// The tokenize share lock store is implemented by keying on the account address
// and storing a timestamp as the value. The timestamp is empty when the lock is
// set and gets populated with the unlock completion time once the unlock has started
func (k Keeper) AddTokenizeSharesLock(ctx context.Context, address sdk.AccAddress) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTokenizeSharesLockKey(address)
	err := store.Set(key, sdk.FormatTimeBytes(time.Time{}))
	if err != nil {
		panic(err)
	}
}

// Removes the tokenize share lock for an account to enable tokenizing shares
func (k Keeper) RemoveTokenizeSharesLock(ctx context.Context, address sdk.AccAddress) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTokenizeSharesLockKey(address)
	err := store.Delete(key)
	if err != nil {
		panic(err)
	}
}

// Updates the timestamp associated with a lock to the time at which the lock expires
func (k Keeper) SetTokenizeSharesUnlockTime(ctx context.Context, address sdk.AccAddress, completionTime time.Time) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTokenizeSharesLockKey(address)
	err := store.Set(key, sdk.FormatTimeBytes(completionTime))
	if err != nil {
		panic(err)
	}
}

// Checks if there is currently a tokenize share lock for a given account
// Returns the status indicating whether the account is locked, unlocked,
// or as a lock expiring. If the lock is expiring, the expiration time is returned
func (k Keeper) GetTokenizeSharesLock(ctx context.Context, address sdk.AccAddress) (status types.TokenizeShareLockStatus, unlockTime time.Time) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTokenizeSharesLockKey(address)
	bz, err := store.Get(key)
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return types.TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED, time.Time{}
	}
	unlockTime, err = sdk.ParseTimeBytes(bz)
	if err != nil {
		panic(err)
	}
	if unlockTime.IsZero() {
		return types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED, time.Time{}
	}
	return types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING, unlockTime
}

// Returns all tokenize share locks
func (k Keeper) GetAllTokenizeSharesLocks(ctx context.Context) (tokenizeShareLocks []types.TokenizeShareLock) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))

	iterator := storetypes.KVStorePrefixIterator(store, types.TokenizeSharesLockPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		addressBz := iterator.Key()[2:] // remove prefix bytes and address length
		unlockTime, err := sdk.ParseTimeBytes(iterator.Value())
		if err != nil {
			panic(err)
		}

		var status types.TokenizeShareLockStatus
		if unlockTime.IsZero() {
			status = types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED
		} else {
			status = types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING
		}

		bechPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
		lock := types.TokenizeShareLock{
			Address:        sdk.MustBech32ifyAddressBytes(bechPrefix, addressBz),
			Status:         status.String(),
			CompletionTime: unlockTime,
		}

		tokenizeShareLocks = append(tokenizeShareLocks, lock)
	}

	return tokenizeShareLocks
}

// Stores a list of addresses pending tokenize share unlocking at the same time
func (k Keeper) SetPendingTokenizeShareAuthorizations(ctx context.Context, completionTime time.Time, authorizations types.PendingTokenizeShareAuthorizations) {
	store := k.storeService.OpenKVStore(ctx)
	timeKey := types.GetTokenizeShareAuthorizationTimeKey(completionTime)
	bz := k.cdc.MustMarshal(&authorizations)
	err := store.Set(timeKey, bz)
	if err != nil {
		panic(err)
	}
}

// Returns a list of addresses pending tokenize share unlocking at the same time
func (k Keeper) GetPendingTokenizeShareAuthorizations(ctx context.Context, completionTime time.Time) types.PendingTokenizeShareAuthorizations {
	store := k.storeService.OpenKVStore(ctx)

	timeKey := types.GetTokenizeShareAuthorizationTimeKey(completionTime)
	bz, err := store.Get(timeKey)
	if err != nil {
		panic(err)
	}

	authorizations := types.PendingTokenizeShareAuthorizations{Addresses: []string{}}
	if len(bz) == 0 {
		return authorizations
	}
	k.cdc.MustUnmarshal(bz, &authorizations)

	return authorizations
}

// Inserts the address into a queue where it will sit for 1 unbonding period
// before the tokenize share lock is removed
// Returns the completion time
func (k Keeper) QueueTokenizeSharesAuthorization(ctx context.Context, address sdk.AccAddress) (time.Time, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	params, err := k.GetParams(ctx)
	if err != nil {
		return blockTime, err
	}

	completionTime := blockTime.Add(params.UnbondingTime)

	// Append the address to the list of addresses that also unlock at this time
	authorizations := k.GetPendingTokenizeShareAuthorizations(ctx, completionTime)
	authorizations.Addresses = append(authorizations.Addresses, address.String())

	k.SetPendingTokenizeShareAuthorizations(ctx, completionTime, authorizations)
	k.SetTokenizeSharesUnlockTime(ctx, address, completionTime)

	return completionTime, nil
}

// Cancels a pending tokenize share authorization by removing the lock from the queue
func (k Keeper) CancelTokenizeShareLockExpiration(ctx context.Context, address sdk.AccAddress, completionTime time.Time) {
	authorizations := k.GetPendingTokenizeShareAuthorizations(ctx, completionTime)

	updatedAddresses := []string{}
	for _, expiringAddress := range authorizations.Addresses {
		if address.String() != expiringAddress {
			updatedAddresses = append(updatedAddresses, expiringAddress)
		}
	}

	authorizations.Addresses = updatedAddresses
	k.SetPendingTokenizeShareAuthorizations(ctx, completionTime, authorizations)
}

// Unlocks all queued tokenize share authorizations that have matured
// (i.e. have waited the full unbonding period)
func (k Keeper) RemoveExpiredTokenizeShareLocks(ctx context.Context, blockTime time.Time) ([]string, error) {
	store := k.storeService.OpenKVStore(ctx)

	// iterators all time slices from time 0 until the current block time
	prefixEnd := storetypes.InclusiveEndBytes(types.GetTokenizeShareAuthorizationTimeKey(blockTime))
	iterator, err := store.Iterator(types.TokenizeSharesUnlockQueuePrefix, prefixEnd)
	if err != nil {
		return []string{}, err
	}

	defer iterator.Close()

	// collect all unlocked addresses
	unlockedAddresses := []string{}
	keys := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		authorizations := types.PendingTokenizeShareAuthorizations{}
		k.cdc.MustUnmarshal(iterator.Value(), &authorizations)
		unlockedAddresses = append(unlockedAddresses, authorizations.Addresses...)
		keys = append(keys, iterator.Key())
	}

	// delete unlocked addresses keys
	for _, k := range keys {
		err := store.Delete(k)
		if err != nil {
			panic(err)
		}
	}

	// remove the lock from each unlocked address
	for _, unlockedAddress := range unlockedAddresses {
		unlockedAddr, err := k.authKeeper.AddressCodec().StringToBytes(unlockedAddress)
		if err != nil {
			return unlockedAddresses, err
		}
		k.RemoveTokenizeSharesLock(ctx, unlockedAddr)
	}

	return unlockedAddresses, nil
}

// Calculates and sets the global liquid staked tokens and liquid shares by validator
// The totals are determined by looping each delegation record and summing the stake
// if the delegator has a 32-length address. Checking for a 32-length address will capture
// ICA accounts, as well as tokenized delegations which are owned by module accounts
// under the hood
// This function must be called in the upgrade handler which onboards LSM
func (k Keeper) RefreshTotalLiquidStaked(ctx context.Context) error {
	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return err
	}

	// First reset each validator's liquid shares to 0
	for _, validator := range validators {
		validator.LiquidShares = math.LegacyZeroDec()
		err = k.SetValidator(ctx, validator)
		if err != nil {
			return err
		}
	}

	delegations, err := k.GetAllDelegations(ctx)
	if err != nil {
		return err
	}

	// Sum up the total liquid tokens and increment each validator's liquid shares
	totalLiquidStakedTokens := math.ZeroInt()
	for _, delegation := range delegations {
		delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(delegation.DelegatorAddress)
		if err != nil {
			return err
		}

		// If the delegator is either an ICA account or a tokenize share module account,
		// the delegation should be considered to be associated with liquid staking
		// Consequently, the global number of liquid staked tokens, and the total
		// liquid shares on the validator should be incremented
		if k.DelegatorIsLiquidStaker(delegatorAddress) {
			validatorAddress, err := k.ValidatorAddressCodec().StringToBytes(delegation.ValidatorAddress)
			if err != nil {
				return err
			}
			validator, err := k.GetValidator(ctx, validatorAddress)
			if err != nil {
				return err
			}

			liquidShares := delegation.Shares
			liquidTokens := validator.TokensFromShares(liquidShares).TruncateInt()

			validator.LiquidShares = validator.LiquidShares.Add(liquidShares)
			err = k.SetValidator(ctx, validator)
			if err != nil {
				return err
			}

			totalLiquidStakedTokens = totalLiquidStakedTokens.Add(liquidTokens)
		}
	}

	k.SetTotalLiquidStakedTokens(ctx, totalLiquidStakedTokens)

	return nil
}

// CheckVestedDelegationInVestingAccount verifies whether the provided vesting account
// holds a vested delegation for an equal or greater amount of the specified coin
// at the given block time.
//
// Note that this function facilitates a specific use-case in the LSM module for tokenizing vested delegations.
// For more details, see https://github.com/cosmos/gaia/issues/2877.
func CheckVestedDelegationInVestingAccount(account vesting.VestingAccount, blockTime time.Time, coin sdk.Coin) bool {
	// Get the vesting coins at the current block time
	vestingAmount := account.GetVestingCoins(blockTime).AmountOf(coin.Denom)

	// Note that the "DelegatedVesting" and "DelegatedFree" values
	// were computed during the last delegation or undelegation operation
	delVestingAmount := account.GetDelegatedVesting().AmountOf(coin.Denom)
	delVested := account.GetDelegatedFree()

	// Calculate the new vested delegated coins
	x := math.MinInt(vestingAmount.Sub(delVestingAmount), math.ZeroInt())

	// Add the newly vested delegated coins to the existing delegated vested amount
	if !x.IsZero() {
		delVested = delVested.Add(sdk.NewCoin(coin.Denom, x.Abs()))
	}

	// Check if the total delegated vested amount is greater than or equal to the specified coin amount
	return delVested.AmountOf(coin.Denom).GTE(coin.Amount)
}
