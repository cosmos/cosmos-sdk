package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetValidatorSlashingPeriod(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t)
	addr := sdk.ValAddress(addrs[0])
	height := int64(5)
	require.Panics(t, func() { keeper.getValidatorSlashingPeriodForHeight(ctx, addr, height) })
	newPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: addr,
		StartHeight:   height,
		EndHeight:     height + 10,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	keeper.setValidatorSlashingPeriod(ctx, newPeriod)
	// Get at start height
	retrieved := keeper.getValidatorSlashingPeriodForHeight(ctx, addr, height)
	require.Equal(t, addr, retrieved.ValidatorAddr)
	// Get before start height
	retrieved = keeper.getValidatorSlashingPeriodForHeight(ctx, addr, int64(0))
	require.Equal(t, addr, retrieved.ValidatorAddr)
	// Get after start height (panic)
	require.Panics(t, func() { keeper.getValidatorSlashingPeriodForHeight(ctx, addr, int64(6)) })
	// Get after end height (panic)
	newPeriod.EndHeight = int64(4)
	keeper.setValidatorSlashingPeriod(ctx, newPeriod)
	require.Panics(t, func() { keeper.getValidatorSlashingPeriodForHeight(ctx, addr, height) })
}

func TestValidatorSlashingPeriodCap(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t)
	addr := sdk.ValAddress(addrs[0])
	height := int64(5)
	newPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: addr,
		StartHeight:   height,
		EndHeight:     height + 10,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	keeper.setValidatorSlashingPeriod(ctx, newPeriod)
	half := sdk.NewDec(1).Quo(sdk.NewDec(2))
	// First slash should be full
	fractionA := keeper.capBySlashingPeriod(ctx, addr, half, height)
	require.True(t, fractionA.Equal(half))
	// Second slash should be capped
	fractionB := keeper.capBySlashingPeriod(ctx, addr, half, height)
	require.True(t, fractionB.Equal(sdk.ZeroDec()))
}
