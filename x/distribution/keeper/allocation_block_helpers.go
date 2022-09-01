package keeper

import (
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
func (k Keeper) GetTotalBlacklistedPower(ctx sdk.Context, valAddr string) (int64, int64, error) {

	blacklistedDelAddrs := k.GetParams(ctx).NoRewardsDelegatorAddresses
	val, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		// something went wrong
		k.Logger(ctx).Error("failed to parse delegator address %s, err %s", val, err.Error())
		return 0, 0, err
	}
	valObj := k.stakingKeeper.Validator(ctx, val)
	valTotPower := sdk.TokensToConsensusPower(valObj.GetTokens(), sdk.DefaultPowerReduction)

	valBlacklistedPower := int64(0)
	for _, delAddr := range blacklistedDelAddrs {
		// there is a check in params.go that prevents invalid addresses from being added
		// so this check should never error
		del, err := sdk.AccAddressFromBech32(delAddr)
		if err != nil {
			// something went wrong
			k.Logger(ctx).Error("failed to parse delegator address %s, err %s", delAddr, err.Error())
			return 0, 0, err
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
	return valTotPower, valBlacklistedPower, nil
}

// function to get totalBlacklistedPowerShare and taintedValsBlacklistedPowerShare
func (k Keeper) GetValsBlacklistedPowerShare(ctx sdk.Context) (totalBlacklistedPower sdk.Dec, blacklistedPowerShareByValidator []types.ValidatorBlacklistedPower) {
	vals := k.GetAllValidators(ctx)
	totalBlacklistedPower = sdk.ZeroDec()
	// runtime is n*m, where n is len(valAddrs) and m is len(blacklistedDelAddrs)
	// in practice, we'd expect n ~= 150 and m ~= 100
	for _, valAddr := range vals {
		// update validator stats
		valPower, valBlacklistedPower, err := k.GetTotalBlacklistedPower(ctx, valAddr)
		if err != nil {
			// something went wrong
			continue
		}
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
