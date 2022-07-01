package keeper

import (
	"fmt"
	"sort"
	"strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// This function decides when to start the auto-delegation
func (k Keeper) isTimeToStartAutoDelegation(ctx sdk.Context) bool {

	// For sake of simplicity we just consider a constant number
	// This can be configured based on the number of auto-delegation requests later on via a linear formula

	const intervalParameter = 10
	return ctx.BlockHeight()%intervalParameter == 0
}

// AutoDelegationBeginBlocker processes the auto delegations
// In order to keep the overhead down, we process one delegation per block
func (k Keeper) AutoDelegationBeginBlocker(ctx sdk.Context) {

	hp := k.GetHeadPointer(ctx)
	if hp.FirstRecord == "" {
		// No auto delegations set
		return
	}

	if hp.CurrentRecord == "" {
		// We hit the end of the list, let's go back to its beginning
		hp.CurrentRecord = hp.FirstRecord

		// When we hit the end of the list, we need to wait for a particular time to start the auto-delegation again
		if !k.isTimeToStartAutoDelegation(ctx) {
			return
		}
		ctx.Logger().Info("Auto delegation started")
	}

	delAddrStr := hp.CurrentRecord
	delAddr, err := sdk.AccAddressFromBech32(delAddrStr)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Auto delegation | account address %s error: %v", delAddrStr, err))
		return
	}

	currentRecord, found := k.GetAutoDelegation(ctx, delAddr)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Auto delegation | retrieving info failed for %s error: %v", delAddrStr, err))
		return
	}

	// Let's move the head pointer to the next record on exit
	defer func() {
		hp.CurrentRecord = currentRecord.NextRecord
		k.SetHeadPointer(ctx, hp)
		if hp.CurrentRecord == "" {
			ctx.Logger().Info("Auto delegation ended")
		}
	}()

	/*-----------*/

	accTokens := make(map[string]sdkmath.Int)

	balances := k.bankKeeper.GetAllBalances(ctx, delAddr)
	if len(balances) == 0 {
		return
	}
	accDenomsSortedList := make([]string, 0, len(balances))

	for i := range balances {
		accTokens[balances[i].Denom] = balances[i].Amount
		accDenomsSortedList = append(accDenomsSortedList, balances[i].Denom)
	}

	// We need to have the denoms sorted in order to be deterministic
	sort.Strings(accDenomsSortedList)

	/*-----------*/

	res, err := k.DelegationTotalRewards(sdk.UnwrapSDKContext(ctx),
		&types.QueryDelegationTotalRewardsRequest{
			DelegatorAddress: delAddrStr,
		},
	)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Auto delegation | TotalRewards of address %s error: %v", delAddrStr, err))
		return
	}

	// We need to find a stateless way to choose to which validator delegate the rewards
	// One way could be get the list of delegated validators and choose the minimum delegation
	// But that requires another query to the state, we can assume that the least reward,
	// could be potentially the validator with the minimum delegated tokens,
	// however it works well if all the delegated validators have the same commission rate, for sake of performance we will go with it
	type validatorRewardMap struct {
		Amount  sdkmath.Int
		Address string
	}
	validatorsToDelegate := make(map[string]validatorRewardMap)

	if len(res.Rewards) == 0 {
		// Empty delegated validator list
		return
	}
	// Find the validators of each denom with min reward
	for i := range res.Rewards {
		for r := range res.Rewards[i].Reward {
			if validatorsToDelegate[res.Rewards[i].Reward[r].Denom].Address == "" {
				validatorsToDelegate[res.Rewards[i].Reward[r].Denom] = validatorRewardMap{
					Address: res.Rewards[i].ValidatorAddress,
					Amount:  res.Rewards[i].Reward[r].Amount.RoundInt(),
				}
			} else if res.Rewards[i].Reward[r].Amount.RoundInt().LT(validatorsToDelegate[res.Rewards[i].Reward[r].Denom].Amount) {
				validatorsToDelegate[res.Rewards[i].Reward[r].Denom] = validatorRewardMap{
					Address: res.Rewards[i].ValidatorAddress,
					Amount:  res.Rewards[i].Reward[r].Amount.RoundInt(),
				}
			}
		}
	}

	/*-----------*/

	for i := range res.Total {
		//TODO: Not sure if it is ok to truncate it or we need to do something with its leftover
		accTokens[res.Total[i].Denom] = accTokens[res.Total[i].Denom].Add(res.Total[i].Amount.TruncateInt())
	}

	/*-----------*/

	alreadyWithdrawn := false
	for i := range accDenomsSortedList {
		denom := accDenomsSortedList[i]
		if accTokens[denom].GT(currentRecord.MinBalance.AmountOf(denom)) {
			// Now let's withdraw the tokens and delegate them
			if !alreadyWithdrawn {
				// Since withdraw is done for all Denoms, let's do it only once
				for v := range res.Rewards {
					valAddr, err := sdk.ValAddressFromBech32(res.Rewards[v].ValidatorAddress)
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("Auto delegation | validator address %s conversion error: %v", res.Rewards[v].ValidatorAddress, err))
						continue
					}
					_, err = k.WithdrawDelegationRewards(ctx, delAddr, valAddr)
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("Auto delegation | withdraw from validator %s failed, error: %v", res.Rewards[v].ValidatorAddress, err))
						continue
					}
				}
			}
			alreadyWithdrawn = true

			//---------//

			// Delegate again the remaining balance
			amountToDelegate := accTokens[denom].Sub(currentRecord.MinBalance.AmountOf(denom))
			//TODO: we need to calculate (estimate) the gas consumption and charge the account here

			if amountToDelegate.GT(sdk.NewInt(0)) { // Just a safety check, instead of zero we can consider a minimum threshold

				valAddrStr := validatorsToDelegate[denom].Address
				valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Auto delegation | validator address %s conversion error: %v", valAddrStr, err))
					return
				}

				validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
				if !found {
					ctx.Logger().Error(fmt.Sprintf("Auto delegation | validator not found: %s", valAddrStr))
					return
				}
				newShares, err := k.stakingKeeper.Delegate(ctx, delAddr, amountToDelegate, stakingtypes.Unbonded, validator, true)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Auto delegation | delegate error: %v \tvalidator: %s \tdelegator: %s", err, valAddrStr, delAddrStr))
					return
				}

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(sdk.EventTypeMessage,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
						sdk.NewAttribute(sdk.AttributeKeyAction, types.AutoDelegationEventKey),
						sdk.NewAttribute(types.ValidatorAddressEventKey, valAddrStr),
						sdk.NewAttribute(types.DelegatorAddressEventKey, delAddrStr),
						sdk.NewAttribute(types.AmountEventKey, strconv.FormatFloat(newShares.MustFloat64(), 'f', 16, 64)),
					),
				)
			}
		}
	}

	/*-----------*/

}

// Set/update the auto delegation info + it updates the head pointer and linked pointers accordingly
func (k Keeper) SetAutoDelegation(ctx sdk.Context, inputRecord types.AutoDelegation) error {

	delAddr, err := sdk.AccAddressFromBech32(inputRecord.DelegatorAddress)
	if err != nil {
		return err
	}

	rec, found := k.GetAutoDelegation(ctx, delAddr)
	if found {
		// Update the existing record
		rec.MinBalance = inputRecord.MinBalance
		k.storeAutoDelegation(ctx, delAddr, rec)
		return nil
	}

	// Add a new record to the head of the list

	hp := k.GetHeadPointer(ctx)

	if hp.FirstRecord != "" {
		firstRecordAddr, err := sdk.AccAddressFromBech32(hp.FirstRecord)
		if err != nil {
			return err
		}
		firstRecord, found := k.GetAutoDelegation(ctx, firstRecordAddr)
		if found {
			firstRecord.PrevRecord = inputRecord.DelegatorAddress
			k.storeAutoDelegation(ctx, firstRecordAddr, firstRecord)
		}
	}

	inputRecord.NextRecord = hp.FirstRecord

	hp.FirstRecord = inputRecord.DelegatorAddress
	k.SetHeadPointer(ctx, hp)

	k.storeAutoDelegation(ctx, delAddr, inputRecord)

	return nil
}

// Stores the auto delegation info as it comes in
func (k Keeper) storeAutoDelegation(ctx sdk.Context, delAddr sdk.AccAddress, autoDelegation types.AutoDelegation) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&autoDelegation)
	store.Set(types.GetDelegatorAutoDelegationKey(delAddr), b)
}

// Unset (remove) the auto delegation info
func (k Keeper) UnSetAutoDelegation(ctx sdk.Context, delAddr sdk.AccAddress) error {

	record, found := k.GetAutoDelegation(ctx, delAddr)
	if !found {
		return nil
	}

	// Check if there is a next record, if so we update it accordingly
	if record.NextRecord != "" {
		nextRecAddr, err := sdk.AccAddressFromBech32(record.NextRecord)
		if err != nil {
			return err
		}
		nextRec, found := k.GetAutoDelegation(ctx, nextRecAddr)
		if found {
			nextRec.PrevRecord = record.PrevRecord
			k.storeAutoDelegation(ctx, nextRecAddr, nextRec)
		}
	}

	if record.PrevRecord != "" {
		prevRecAddr, err := sdk.AccAddressFromBech32(record.PrevRecord)
		if err != nil {
			return err
		}
		prevRec, found := k.GetAutoDelegation(ctx, prevRecAddr)
		if found {
			prevRec.NextRecord = record.NextRecord
			k.storeAutoDelegation(ctx, prevRecAddr, prevRec)
		}
	}

	/*---------*/

	// Check if the current head pointer is pointing to this record, move it to the next one
	hp := k.GetHeadPointer(ctx)
	hpModified := false
	if hp.CurrentRecord == record.DelegatorAddress {
		hp.CurrentRecord = record.NextRecord
		hpModified = true
	}
	if hp.FirstRecord == record.DelegatorAddress {
		hp.FirstRecord = record.NextRecord
		hpModified = true
	}
	if hpModified {
		k.SetHeadPointer(ctx, hp)
	}

	/*---------*/

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDelegatorAutoDelegationKey(delAddr))

	return nil
}

// GetAutoDelegation searches and returns the auto delegation on the given delegator address
func (k Keeper) GetAutoDelegation(ctx sdk.Context, delAddr sdk.AccAddress) (record types.AutoDelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	index := types.GetDelegatorAutoDelegationKey(delAddr)

	b := store.Get(index)
	if b == nil {
		return record, false
	}

	k.cdc.MustUnmarshal(b, &record)
	return record, true
}

// GetHeadPointer returns the Head Pointer
func (k Keeper) GetHeadPointer(ctx sdk.Context) (headPointer types.HeadPointer) {

	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetHeadPointerAutoDelegationPrefix())
	if b == nil {
		return headPointer
	}
	k.cdc.MustUnmarshal(b, &headPointer)
	return headPointer
}

// SetHeadPointer sets the head pointer
func (k Keeper) SetHeadPointer(ctx sdk.Context, headPointer types.HeadPointer) {

	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&headPointer)
	store.Set(types.GetHeadPointerAutoDelegationPrefix(), b)
}

// PrintAllAutoDelegations is useful for development and debugging
func (k Keeper) PrintAllAutoDelegations(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegatorAutoDelegationPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.AutoDelegation
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		fmt.Printf("\n----------------------------\n\n%#v\n", val)
	}
}
