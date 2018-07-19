package gas

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

func TestGasAnteHandler(t *testing.T) {
	ctx, store, _ := store.DefaultTestComponents(t)

	cases := []sdk.GasConfig{
		sdk.GasConfig{3, 3, 3, 3, 3, 3, 3, 3},
		sdk.GasConfig{20, 20, 2, 20, 20, 10, 20, 2},
		sdk.GasConfig{50, 80, 20, 40, 55, 15, 25, 5},
	}

	ante := NewAnteHandler(store)

	var abort bool
	ctx, _, abort = ante(ctx, nil)
	require.False(t, abort)
	require.Equal(t, sdk.DefaultKVGasConfig(), ctx.KVGasConfig())
	require.Equal(t, sdk.DefaultTransientGasConfig(), ctx.TransientGasConfig())

	for _, tc := range cases {
		store.Set(ctx, kvStoreKey, tc)
		store.Set(ctx, transientStoreKey, tc)
		ctx, _, abort = ante(ctx, nil)
		require.False(t, abort)
		require.Equal(t, tc, ctx.KVGasConfig())
		require.Equal(t, tc, ctx.TransientGasConfig())
	}
}
