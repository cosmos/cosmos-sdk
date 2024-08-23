package implementation

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/core/transaction"
)

func TestMakeAccountContext(t *testing.T) {
	originalContext := coretesting.Context()
	storeService := coretesting.KVStoreService(originalContext, "test")
	accountAddr := []byte("accountAddr")
	sender := []byte("sender")
	sb := collections.NewSchemaBuilderFromAccessor(openKVStore)

	accountCtx := MakeAccountContext(originalContext, storeService, 1, accountAddr, sender, nil, nil, nil)

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
	// ensure calling ExecModuleUntyped works
	accountCtx = MakeAccountContext(originalContext, storeService, 1, []byte("legit-exec-module-untyped"), []byte("invoker"), nil, func(ctx context.Context, sender []byte, msg transaction.Msg) (transaction.Msg, error) {
		require.Equal(t, originalContext, ctx)
		return &types.StringValue{Value: "module exec untyped was called"}, nil
	}, nil)

	respUntyped, err := ExecModule(accountCtx, &types.UInt64Value{Value: 1000})
	require.NoError(t, err)
	require.True(t, Equal(&types.StringValue{Value: "module exec untyped was called"}, respUntyped))

	// ensure calling QueryModule works, also by setting everything else communication related to nil
	// we can guarantee that exec paths do not impact query paths.
	accountCtx = MakeAccountContext(originalContext, storeService, 1, nil, nil, nil, nil, func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		require.Equal(t, originalContext, ctx)
		return &types.StringValue{Value: "module query was called"}, nil
	})

	resp, err := QueryModule(accountCtx, &types.UInt64Value{Value: 1000})
	require.NoError(t, err)
	require.True(t, Equal(&types.StringValue{Value: "module query was called"}, resp))
}
