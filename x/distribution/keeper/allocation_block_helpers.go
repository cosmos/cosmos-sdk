package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetAllValidators(ctx sdk.Context) (validatorAddresses []string) {
	k.stakingKeeper.IterateValidators(
		ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
			// Only consider active validators; inactive validators can't have signed the last block (CHECK ASSUMPTION)
			if val.IsBonded() {
				validatorAddresses = append(validatorAddresses, val.GetOperator().String())
			}
			return false
		},
	)
	return
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

// get a validator's total blacklisted delegation power
// 		returns (totalPower, blacklistedPower)
func (k Keeper) GetTotalBlacklistedPower(ctx sdk.Context, valAddr string) (int64, int64) {

	blacklistedDelAddrs := k.GetParams(ctx).NoRewardsDelegatorAddresses
	fmt.Println("blacklistedDelAddrs", blacklistedDelAddrs)
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

// function to get totalBlacklistedPowerShare and taintedValsBlacklistedPowerShare
func (k Keeper) GetValsBlacklistedPowerShare(ctx sdk.Context) (totalBlacklistedPower sdk.Dec, blacklistedPowerShareByValidator []types.ValidatorBlacklistedPower) {
	vals := k.GetAllValidators(ctx)
	fmt.Println("GetValsBlacklistedPowerShare vals", vals)
	// runtime is n*m, where n is len(valAddrs) and m is len(blacklistedDelAddrs)
	// in practice, we'd expect n ~= 150 and m ~= 100
	for _, valAddr := range vals {
		// update validator stats
		fmt.Println("valAddr", valAddr)
		valPower, valBlacklistedPower := k.GetTotalBlacklistedPower(ctx, valAddr)
		fmt.Println("valPower", valPower)
		fmt.Println("valBlacklistedPower", valBlacklistedPower)
		valBlacklistedPowerShare := sdk.NewDec(valBlacklistedPower).Quo(sdk.NewDec(valPower))
		blacklistedPowerShareByValidator = append(blacklistedPowerShareByValidator, types.ValidatorBlacklistedPower{
			ValidatorAddress:      valAddr,
			BlacklistedPowerShare: valBlacklistedPowerShare,
		})
		// update summary stats
		totalBlacklistedPower = totalBlacklistedPower.Add(sdk.NewDec(valBlacklistedPower))
	}
	return totalBlacklistedPower, blacklistedPowerShareByValidator
}

func (k Keeper) GetBlacklistedPowerShareByValidator(ctx sdk.Context, validatorBlacklistedPowers []types.ValidatorBlacklistedPower) (blacklistedPowerShareByValidator map[string]sdk.Dec) {
	blacklistedPowerShareByValidator = make(map[string]sdk.Dec)
	for _, valBlacklistedPower := range validatorBlacklistedPowers {
		blacklistedPowerShareByValidator[valBlacklistedPower.ValidatorAddress] = valBlacklistedPower.BlacklistedPowerShare
	}
	return blacklistedPowerShareByValidator
}
