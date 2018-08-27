package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestHookOnValidatorBonded(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t)
	addr := sdk.ValAddress(addrs[0])
	keeper.onValidatorBonded(ctx, addr)
	period := keeper.getValidatorSlashingPeriodForHeight(ctx, addr, ctx.BlockHeight())
	require.Equal(t, ValidatorSlashingPeriod{addr, ctx.BlockHeight(), 0, sdk.ZeroDec()}, period)
}

func TestHookOnValidatorUnbonded(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t)
	addr := sdk.ValAddress(addrs[0])
	keeper.onValidatorBonded(ctx, addr)
	keeper.onValidatorUnbonded(ctx, addr)
	period := keeper.getValidatorSlashingPeriodForHeight(ctx, addr, ctx.BlockHeight())
	require.Equal(t, ValidatorSlashingPeriod{addr, ctx.BlockHeight(), ctx.BlockHeight(), sdk.ZeroDec()}, period)
}
