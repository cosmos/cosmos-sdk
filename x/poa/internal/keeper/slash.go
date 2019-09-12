package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Slash a validator for an infraction committed at a known height
// Find the validators weight at that height and burn the specified slashFactor
// of it
//
// CONTRACT:
//    slashFactor is non-negative
// CONTRACT:
//    Infraction was committed equal to or less than an unbonding period in the past,
// CONTRACT:
//    Slash will not slash unbonded validators (for the above reason)
// CONTRACT:
//    Infraction was committed at the current height or at a past height,
//    not at a height in the future

func (k Keeper) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdk.Dec) {
	logger := k.Logger(ctx)

	if slashFactor.LT(sdk.ZeroDec()) {
		panic(fmt.Errorf("attempted to slash with a negative slash factor: %v", slashFactor))
	}

	// Amount of slashing = slash slashFactor * power at time of infraction
	amount := sdk.TokensFromConsensusPower(power) // reuse of a function, but it relates to weight
	slashAmountDec := amount.ToDec().Mul(slashFactor)
	slashAmount := slashAmountDec.TruncateInt()

	// ref https://github.com/cosmos/cosmos-sdk/issues/1348

	validator, found := k.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		// If not found, the validator must have been overslashed and removed - so we don't need to do anything
		// NOTE:  Correctness dependent on invariant that unbonding delegations / redelegations must also have been completely
		//        slashed in this case - which we don't explicitly check, but should be true.
		// Log the slash attempt for future reference (maybe we should tag it too)
		logger.Error(fmt.Sprintf(
			"WARNING: Ignored attempt to slash a nonexistent validator with address %s, we recommend you investigate immediately",
			consAddr))
		return
	}

	// should not be slashing an unbonded validator
	if validator.IsUnbonded() {
		panic(fmt.Sprintf("should not be slashing unbonded validator: %s", validator.GetOperator()))
	}

	operatorAddress := validator.GetOperator()

	// call the before-modification hook
	k.BeforeValidatorModified(ctx, operatorAddress)

	// Track remaining slash amount for the validator
	// This will decrease when we slash unbondings and
	// redelegations, as that stake has since unbonded
	remainingSlashAmount := slashAmount

	switch {
	case infractionHeight > ctx.BlockHeight():

		// Can't slash infractions in the future
		panic(fmt.Sprintf(
			"impossible attempt to slash future infraction at height %d but we are at height %d",
			infractionHeight, ctx.BlockHeight()))

	case infractionHeight == ctx.BlockHeight():

		// Special-case slash at current height for efficiency - we don't need to look through unbonding delegations or redelegations
		logger.Info(fmt.Sprintf(
			"slashing at current height %d, not scanning unbonding delegations & redelegations",
			infractionHeight))

	case infractionHeight < ctx.BlockHeight():

		// cannot decrease balance below zero
		weightRemoval := sdk.MinInt(remainingSlashAmount, validator.Weight)
		weightRemoval = sdk.MaxInt(weightRemoval, sdk.ZeroInt()) // defensive.

		// we need to calculate the *effective* slash fraction for distribution
		if validator.Weight.GT(sdk.ZeroInt()) {
			effectiveFraction := weightRemoval.ToDec().QuoRoundUp(validator.Weight.ToDec())
			// possible if power has changed
			if effectiveFraction.GT(sdk.OneDec()) {
				effectiveFraction = sdk.OneDec()
			}
			// call the before-slashed hook
			k.BeforeValidatorSlashed(ctx, operatorAddress, effectiveFraction)
		}

		// switch validator.GetStatus() {
		// case sdk.Bonded:
		// 	if err := k.burnBondedTokens(ctx, tokensToBurn); err != nil {
		// 		panic(err)
		// 	}
		// case sdk.Unbonding, sdk.Unbonded:
		// 	if err := k.burnNotBondedTokens(ctx, tokensToBurn); err != nil {
		// 		panic(err)
		// 	}
		// default:
		// 	panic("invalid validator status")
		// }

		// Log that a slash occurred!
		logger.Info(fmt.Sprintf(
			"validator %s slashed by slash factor of %s; burned %v tokens",
			validator.GetOperator(), slashFactor.String(), weightRemoval))

		return
	}
}

// jail a validator

func (k Keeper) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	validator := k.mustGetValidatorByConsAddr(ctx, consAddr)
	k.jailValidator(ctx, validator)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("validator %s jailed", consAddr))
	return
}

// unjail a validator
func (k Keeper) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	validator := k.mustGetValidatorByConsAddr(ctx, consAddr)
	k.unjailValidator(ctx, validator)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("validator %s unjailed", consAddr))
	return
}
