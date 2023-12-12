package accounts

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/x/accounts/internal/implementation"
)

func TestGenesis(t *testing.T) {
	k, ctx := newKeeper(t, func(deps implementation.Dependencies) (string, implementation.Account, error) {
		acc, err := NewTestAccount(deps)
		return "test", acc, err
	})
	k.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error { return nil })
	// we init two accounts of the same type

	// we set counter to 10
	_, addr1, err := k.Init(ctx, "test", []byte("sender"), &types.Empty{})
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr1, []byte("sender"), &types.UInt64Value{Value: 10})
	require.NoError(t, err)

	// we set counter to 20
	_, addr2, err := k.Init(ctx, "test", []byte("sender"), &types.Empty{})
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr2, []byte("sender"), &types.UInt64Value{Value: 20})
	require.NoError(t, err)

	// export state
	state, err := k.ExportState(ctx)
	require.NoError(t, err)

	// reset state
	_, ctx = colltest.MockStore()
	err = k.ImportState(ctx, state)
	require.NoError(t, err)

	// if genesis import went fine, we should be able to query the accounts
	// and get the expected values.
	resp, err := k.Query(ctx, addr1, &types.DoubleValue{})
	require.NoError(t, err)
	require.Equal(t, &types.UInt64Value{Value: 10}, resp)

	resp, err = k.Query(ctx, addr2, &types.DoubleValue{})
	require.NoError(t, err)
	require.Equal(t, &types.UInt64Value{Value: 20}, resp)
}
