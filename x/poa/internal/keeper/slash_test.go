package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// setup helper function - creates two validators
func setupHelper(t *testing.T) (sdk.Context, Keeper, types.Params) {

	// setup
	ctx, _, keeper, _ := CreateTestInput(t, false)
	params := keeper.GetParams(ctx)
	numVals := int64(3)

	// add numVals validators
	for i := int64(0); i < numVals; i++ {
		validator := types.NewValidator(addrVals[i], PKs[i], stakingtypes.Description{})
		validator = TestingUpdateValidator(keeper, ctx, validator, true)
		keeper.SetValidatorByConsAddr(ctx, validator)
	}

	return ctx, keeper, params
}

func TestRevocation(t *testing.T) {
	ctx, keeper, _ := setupHelper(t)
	addr := addrVals[0]
	consAddr := sdk.ConsAddress(PKs[0].Address())

	val, ok := keeper.GetValidator(ctx, addr)
	require.True(t, ok)
	require.False(t, val.IsJailed())

	keeper.Jail(ctx, consAddr)
	val, ok = keeper.GetValidator(ctx, addr)
	require.True(t, ok)
	require.True(t, val.IsJailed())

	keeper.Unjail(ctx, consAddr)
	val, ok = keeper.GetValidator(ctx, addr)
	require.True(t, ok)
	require.False(t, val.IsJailed())
}

// tests Slash at a future height (must panic)
func TestSlashAtFutureHeight(t *testing.T) {
	ctx, keeper, _ := setupHelper(t)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	require.Panics(t, func() { keeper.Slash(ctx, consAddr, 1, 10, fraction) })
}

// tests Slash at the current height
// func TestSlashValidatorAtCurrentHeight(t *testing.T) {
// 	ctx, keeper, _ := setupHelper(t)
// 	consAddr := sdk.ConsAddress(PKs[0].Address())
// 	fraction := sdk.NewDecWithPrec(5, 1)

// 	validator, found := keeper.GetValidatorByConsAddr(ctx, consAddr)
// 	require.True(t, found)
// 	keeper.Slash(ctx, consAddr, ctx.BlockHeight(), 10, fraction)

// 	// read updated state
// 	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
// 	require.True(t, found)

// 	// end block
// 	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
// 	fmt.Println("here")
// 	require.Equal(t, 1, len(updates), "cons addr: %v, updates: %v", []byte(consAddr), updates)

// 	validator = keeper.mustGetValidator(ctx, validator.OperatorAddress)
// 	// power decreased
// 	require.Equal(t, int64(5), validator.GetConsensusPower())
// }
