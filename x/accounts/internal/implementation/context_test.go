package implementation

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
)

func TestMakeAccountContext(t *testing.T) {
	storeService, originalContext := colltest.MockStore()
	accountAddr := []byte("accountAddr")
	sender := []byte("sender")
	sb := collections.NewSchemaBuilderFromAccessor(openKVStore)

	accountCtx := MakeAccountContext(originalContext, storeService, 1, accountAddr, sender, nil, nil, nil, nil)

	// ensure whoami
	require.Equal(t, accountAddr, Whoami(accountCtx))
	// ensure sender
	require.Equal(t, sender, Sender(accountCtx))

	ta, err := NewTestAccount(sb)
	require.NoError(t, err)

	impl, err := newImplementation(sb, ta)
	require.NoError(t, err)

	_, err = impl.Execute(accountCtx, &types.UInt64Value{Value: 1000})
	require.NoError(t, err)

	// we want to ensure that the account wrote in the correct prefix.
	// this store is the global x/accounts module store.
	store := storeService.OpenKVStore(originalContext)

	// now we want the value to be store in the following accounts prefix (AccountsStatePrefix + big_endian(acc_number=1) + itemPrefix)
	value, err := store.Get(append(AccountStatePrefix, append(binary.BigEndian.AppendUint64(nil, 1), itemPrefix...)...))
	require.NoError(t, err)
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 3, 232}, value)

	// ensure calling ExecModule works
	accountCtx = MakeAccountContext(originalContext, storeService, 1, []byte("legit-exec-module"), []byte("invoker"), nil, func(ctx context.Context, sender []byte, msg, msgResp ProtoMsg) error {
		// ensure we unwrapped the context when invoking a module call
		require.Equal(t, originalContext, ctx)
		Merge(msgResp, &types.StringValue{Value: "module exec was called"})
		return nil
	}, nil, nil)

	resp, err := ExecModule[types.StringValue](accountCtx, &types.UInt64Value{Value: 1000})
	require.NoError(t, err)
	require.True(t, Equal(&types.StringValue{Value: "module exec was called"}, resp))

	// ensure calling ExecModuleUntyped works
	accountCtx = MakeAccountContext(originalContext, storeService, 1, []byte("legit-exec-module-untyped"), []byte("invoker"), nil, nil, func(ctx context.Context, sender []byte, msg ProtoMsg) (ProtoMsg, error) {
		require.Equal(t, originalContext, ctx)
		return &types.StringValue{Value: "module exec untyped was called"}, nil
	}, nil)

	respUntyped, err := ExecModuleUntyped(accountCtx, &types.UInt64Value{Value: 1000})
	require.NoError(t, err)
	require.True(t, Equal(&types.StringValue{Value: "module exec untyped was called"}, respUntyped))

	// ensure calling QueryModule works, also by setting everything else communication related to nil
	// we can guarantee that exec paths do not impact query paths.
	accountCtx = MakeAccountContext(originalContext, storeService, 1, nil, nil, nil, nil, nil, func(ctx context.Context, req, resp ProtoMsg) error {
		require.Equal(t, originalContext, ctx)
		Merge(resp, &types.StringValue{Value: "module query was called"})
		return nil
	})

	resp, err = QueryModule[types.StringValue](accountCtx, &types.UInt64Value{Value: 1000})
	require.NoError(t, err)
	require.True(t, Equal(&types.StringValue{Value: "module query was called"}, resp))
}
