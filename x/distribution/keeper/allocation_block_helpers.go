package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) unionStrSlices(a, b []string) []string {
	m := make(map[string]bool)
	sort.Strings(a)
	sort.Strings(b)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

// get a delegator's validators, based on grpc_query.go's `DelegatorValidators()`
func (k Keeper) GetDelegatorValidators(ctx sdk.Context, delAddr string) ([]string, error) {
	delAdr, err := sdk.AccAddressFromBech32(delAddr)
	if err != nil {
		return nil, err
	}
	var validators []string

	k.stakingKeeper.IterateDelegations(
		ctx, delAdr,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			validators = append(validators, del.GetValidatorAddr().String())
			return false
		},
	)
	return validators, nil
}

// iterate the blacklisted delegators to gather a list of validators they're delegated to
func (k Keeper) GetTaintedValidators(ctx sdk.Context) []string {
	// get the list of blacklisted delegators
	blacklistedDelAddrs := k.GetParams(ctx).NoRewardsDelegatorAddresses
	// get the list of validators they're delegated to
	taintedVals := []string{}
	for _, delAddr := range blacklistedDelAddrs {

		// can we invoke grpc like this? hacky? unsafe?
		queryValsResp, err := k.GetDelegatorValidators(ctx, delAddr)
		if err != nil {
			panic(err)
		}
		validators := queryValsResp
		taintedVals = k.unionStrSlices(taintedVals, validators)
	}
	k.Logger(ctx).Info(fmt.Sprintf("taintedvals: %#v", taintedVals))
	return taintedVals
}

// get a validator's total blacklisted delegation power
// 		returns (totalPower, blacklistedPower)
func (k Keeper) GetBlacklistedPower(ctx sdk.Context, valAddr string) (int64, int64) {

	blacklistedDelAddrs := k.GetParams(ctx).NoRewardsDelegatorAddresses
	// k.Logger(ctx).Info("Blacklisted delegators", "addrs", blacklistedDelAddrs)
	// get validator
	val, error := sdk.ValAddressFromBech32(valAddr)
	if error != nil {
		// TODO: panic?
		panic(error)
	}
	valObj := k.stakingKeeper.Validator(ctx, val)
	valTotPower := sdk.TokensToConsensusPower(valObj.GetTokens(), sdk.DefaultPowerReduction)

	valBlacklistedPower := int64(0)
	for _, delAddr := range blacklistedDelAddrs {
		// convert delAddrs to dels
		del, err := sdk.AccAddressFromBech32(delAddr)
		if err != nil {
			// TODO: panic?
			panic(err)
		}

		// add the delegation share to total
		delegation := k.stakingKeeper.Delegation(ctx, del, val)
		if delegation != nil {
			// TODO: why does TokensFromShares return a dec, when all tokens are ints? I truncate manually here -- is that safe?
			shares := delegation.GetShares()
			tokens := valObj.TokensFromShares(shares).TruncateInt()
			consPower := sdk.TokensToConsensusPower(tokens, sdk.DefaultPowerReduction)
			valBlacklistedPower = valBlacklistedPower + consPower
		}
	}
	// k.Logger(ctx).Info(fmt.Sprintf("Total valBlacklistedPower is %d", valBlacklistedPower))
	return valTotPower, valBlacklistedPower
}

// helper
func (k Keeper) StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// function to get totalBlacklistedPowerShare and taintedValsBlacklistedPowerShare
func (k Keeper) GetValsBlacklistedPowerShare(ctx sdk.Context) (sdk.Dec, map[string]sdk.Dec, []string) {
	// get the blacklisted validators from the param store
	taintedVals := k.GetTaintedValidators(ctx)
	valsBlacklistedPower := int64(0)
	valsTotalPower := int64(0)
	taintedValsBlacklistedPowerShare := map[string]sdk.Dec{}
	for _, valAddr := range taintedVals {
		valTotalPower, valBlacklistedPower := k.GetBlacklistedPower(ctx, valAddr)
		valBlacklistedPowerShare := sdk.NewDec(valBlacklistedPower).Quo(sdk.NewDec(valTotalPower))
		valsBlacklistedPower += valBlacklistedPower
		valsTotalPower += valTotalPower
		taintedValsBlacklistedPowerShare[valAddr] = valBlacklistedPowerShare
	}
	return sdk.NewDec(valsBlacklistedPower), taintedValsBlacklistedPowerShare, taintedVals
}
