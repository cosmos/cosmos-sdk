package accounts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/x/accounts/internal/implementation"
)

func TestGenesis(t *testing.T) {
	k, ctx := newKeeper(t, implementation.AddAccount("test", NewTestAccount))
	k.queryModuleFunc = func(ctx context.Context, req, resp protoiface.MessageV1) error {
		return nil
	}

	// we init two accounts of the same type

	// we set counter to 10
	_, addr1, err := k.Init(ctx, "test", []byte("sender"), &emptypb.Empty{})
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr1, []byte("sender"), &wrapperspb.UInt64Value{Value: 10})
	require.NoError(t, err)

	// we set counter to 20
	_, addr2, err := k.Init(ctx, "test", []byte("sender"), &emptypb.Empty{})
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr2, []byte("sender"), &wrapperspb.UInt64Value{Value: 20})
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
	resp, err := k.Query(ctx, addr1, &wrapperspb.DoubleValue{})
	require.NoError(t, err)
	require.Equal(t, &wrapperspb.UInt64Value{Value: 10}, resp)

	resp, err = k.Query(ctx, addr2, &wrapperspb.DoubleValue{})
	require.NoError(t, err)
	require.Equal(t, &wrapperspb.UInt64Value{Value: 20}, resp)
}
