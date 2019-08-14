package operations

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)


// RandomValidator returns a random validator given access to the keeper and ctx
func RandomValidator(r *rand.Rand, k keeper.Keeper, ctx sdk.Context) types.Validator {
	vals := k.GetAllValidators(ctx)
	i := r.Intn(len(vals))
	return vals[i]
}

// RandomBondedValidator returns a random bonded validator given access to the keeper and ctx
func RandomBondedValidator(r *rand.Rand, k keeper.Keeper, ctx sdk.Context) types.Validator {
	vals := k.GetBondedValidatorsByPower(ctx)
	i := r.Intn(len(vals))
	return vals[i]
}