package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetValidatorSlashingPeriod(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t, DefaultParams())
	addr := sdk.ConsAddress(addrs[0])
	height := int64(5)
	require.Panics(t, func() { keeper.getValidatorSlashingPeriodForHeight(ctx, addr, height) })
	newPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: addr,
		StartHeight:   height,
		EndHeight:     height + 10,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	keeper.addOrUpdateValidatorSlashingPeriod(ctx, newPeriod)

	// Get at start height
	retrieved := keeper.getValidatorSlashingPeriodForHeight(ctx, addr, height)
	require.Equal(t, newPeriod, retrieved)

	// Get after start height (works)
	retrieved = keeper.getValidatorSlashingPeriodForHeight(ctx, addr, int64(6))
	require.Equal(t, newPeriod, retrieved)

	// Get before start height (panic)
	require.Panics(t, func() { keeper.getValidatorSlashingPeriodForHeight(ctx, addr, int64(0)) })

	// Get after end height (panic)
	newPeriod.EndHeight = int64(4)
	keeper.addOrUpdateValidatorSlashingPeriod(ctx, newPeriod)
	require.Panics(t, func() { keeper.capBySlashingPeriod(ctx, addr, sdk.ZeroDec(), height) })

	// Back to old end height
	newPeriod.EndHeight = height + 10
	keeper.addOrUpdateValidatorSlashingPeriod(ctx, newPeriod)

	// Set a new, later period
	anotherPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: addr,
		StartHeight:   height + 1,
		EndHeight:     height + 11,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	keeper.addOrUpdateValidatorSlashingPeriod(ctx, anotherPeriod)

	// Old period retrieved for prior height
	retrieved = keeper.getValidatorSlashingPeriodForHeight(ctx, addr, height)
	require.Equal(t, newPeriod, retrieved)

	// New period retrieved at new height
	retrieved = keeper.getValidatorSlashingPeriodForHeight(ctx, addr, height+1)
	require.Equal(t, anotherPeriod, retrieved)
}

func TestValidatorSlashingPeriodCap(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t, DefaultParams())
	addr := sdk.ConsAddress(addrs[0])
	height := int64(5)
	newPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: addr,
		StartHeight:   height,
		EndHeight:     height + 10,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	keeper.addOrUpdateValidatorSlashingPeriod(ctx, newPeriod)
	half := sdk.NewDec(1).Quo(sdk.NewDec(2))

	// First slash should be full
	fractionA := keeper.capBySlashingPeriod(ctx, addr, half, height)
	require.True(t, fractionA.Equal(half))

	// Second slash should be capped
	fractionB := keeper.capBySlashingPeriod(ctx, addr, half, height)
	require.True(t, fractionB.Equal(sdk.ZeroDec()))

	// Third slash should be capped to difference
	fractionC := keeper.capBySlashingPeriod(ctx, addr, sdk.OneDec(), height)
	require.True(t, fractionC.Equal(half))
}
