package gas

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/space"
)

func TestGasAnteHandler(t *testing.T) {
	ctx, space, _ := space.DefaultTestComponents(t)

	cases := []sdk.GasConfig{
		sdk.GasConfig{3, 3, 3, 3, 3, 3, 3, 3},
		sdk.GasConfig{20, 20, 2, 20, 20, 10, 20, 2},
		sdk.GasConfig{50, 80, 20, 40, 55, 15, 25, 5},
	}

	ante := NewAnteHandler(space)

	var abort bool
	ctx, _, abort = ante(ctx, nil)
	require.False(t, abort)
	require.Equal(t, sdk.DefaultKVGasConfig(), ctx.KVGasConfig())
	require.Equal(t, sdk.DefaultTransientGasConfig(), ctx.TransientGasConfig())

	for _, tc := range cases {
		space.Set(ctx, kvStoreKey, tc)
		space.Set(ctx, transientStoreKey, tc)
		ctx, _, abort = ante(ctx, nil)
		require.False(t, abort)
		require.Equal(t, tc, ctx.KVGasConfig())
		require.Equal(t, tc, ctx.TransientGasConfig())
	}
}
