package testutil

import (
	"fmt"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func CreateValidator(pk cryptotypes.PubKey) (stakingtypes.Validator, error) {
	valConsAddr := sdk.GetConsAddress(pk)
	val, err := stakingtypes.NewValidator(sdk.ValAddress(valConsAddr), pk, stakingtypes.Description{})
	return val, err
}

func CallCreateValidatorHooks(ctx sdk.Context, k keeper.Keeper, addr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := k.Hooks().AfterValidatorCreated(ctx, valAddr)
	if err != nil {
		return err
	}

	err = k.Hooks().BeforeDelegationCreated(ctx, addr, valAddr)
	if err != nil {
		return err
	}

	k.Hooks().AfterDelegationModified(ctx, addr, valAddr)
	if err != nil {
		return err
	}

	return nil
}

// SlashValidator copies what x/staking Slash does. It should be used for testing only.
// And it must be updated whenever the original function is updated.
func SlashValidator(
	ctx sdk.Context,
	consAddr sdk.ConsAddress,
	infractionHeight int64,
	power int64,
	slashFactor sdk.Dec,
	validator *stakingtypes.Validator,
	distrKeeper *keeper.Keeper,
) math.Int {
	if slashFactor.IsNegative() {
		panic(fmt.Errorf("attempted to slash with a negative slash factor: %v", slashFactor))
	}

	// call the before-modification hook
	distrKeeper.Hooks().BeforeValidatorModified(ctx, validator.GetOperator())

	// we simplify this part, as we won't be able to test redelegations or
	// unbonding delegations
	if infractionHeight != ctx.BlockHeight() {
		// if a new test lands here we might need to update this function to handle redelegations and unbonding
		// or just make it an integration test.
		panic("we can't test any other case here")
	}

	slashAmountDec := sdk.NewDecFromInt(validator.Tokens).Mul(sdk.NewDecWithPrec(5, 1))
	slashAmount := slashAmountDec.TruncateInt()

	// cannot decrease balance below zero
	tokensToBurn := sdk.MinInt(slashAmount, validator.Tokens)
	tokensToBurn = sdk.MaxInt(tokensToBurn, math.ZeroInt()) // defensive.

	// we need to calculate the *effective* slash fraction for distribution
	if validator.Tokens.IsPositive() {
		effectiveFraction := sdk.NewDecFromInt(tokensToBurn).QuoRoundUp(sdk.NewDecFromInt(validator.Tokens))
		// possible if power has changed
		if effectiveFraction.GT(math.LegacyOneDec()) {
			effectiveFraction = math.LegacyOneDec()
		}
		// call the before-slashed hook
		distrKeeper.Hooks().BeforeValidatorSlashed(ctx, validator.GetOperator(), effectiveFraction)
	}
	// Deduct from validator's bonded tokens and update the validator.
	// Burn the slashed tokens from the pool account and decrease the total supply.
	validator.Tokens = validator.Tokens.Sub(tokensToBurn)

	return tokensToBurn
}
