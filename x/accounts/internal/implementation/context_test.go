package implementation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
)

func TestMakeAccountContext(t *testing.T) {
	storeService, originalContext := colltest.MockStore()
	accountAddr := []byte("accountAddr")
	sender := []byte("sender")
	sb := collections.NewSchemaBuilderFromAccessor(OpenKVStore)

	accountCtx := MakeAccountContext(originalContext, storeService, accountAddr, sender, nil, nil, nil)

	// ensure whoami
	require.Equal(t, accountAddr, Whoami(accountCtx))
	// ensure sender
	require.Equal(t, sender, Sender(accountCtx))

	ta, err := NewTestAccount(sb)
	require.NoError(t, err)

	impl, err := NewImplementation(ta)
	require.NoError(t, err)

	_, err = impl.Execute(accountCtx, &wrapperspb.UInt64Value{Value: 1000})
	require.NoError(t, err)

	// we want to ensure that the account wrote in the correct prefix.
	// this store is the global x/accounts module store.
	store := storeService.OpenKVStore(originalContext)

	// now we want the value to be store in the following accounts prefix (AccountsStatePrefix + accountAddr + itemPrefix)
	value, err := store.Get(append(AccountStatePrefix, append(accountAddr, itemPrefix...)...))
	require.NoError(t, err)
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 3, 232}, value)

	// ensure getSenderAccount blocks impersonation
	accountCtx = MakeAccountContext(originalContext, storeService, []byte("impersonator"), []byte("account-invoker"), func(_ proto.Message) ([]byte, error) {
		return []byte("legit-exec-module"), nil
	}, nil, nil)

	_, err = ExecModule[wrapperspb.StringValue](accountCtx, &wrapperspb.UInt64Value{Value: 1000})
	require.ErrorIs(t, err, errUnauthorized)

	// ensure calling ExecModule works
	accountCtx = MakeAccountContext(originalContext, storeService, []byte("legit-exec-module"), []byte("invoker"), func(_ proto.Message) ([]byte, error) {
		return []byte("legit-exec-module"), nil
	}, func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		// ensure we unwrapped the context when invoking a module call
		require.Equal(t, originalContext, ctx)
		return wrapperspb.String("module exec was called"), nil
	}, nil)

	resp, err := ExecModule[wrapperspb.StringValue](accountCtx, &wrapperspb.UInt64Value{Value: 1000})
	require.NoError(t, err)
	require.True(t, proto.Equal(wrapperspb.String("module exec was called"), resp))

	// ensure calling QueryModule works, also by setting everything else communication related to nil
	// we can guarantee that exec paths do not impact query paths.
	accountCtx = MakeAccountContext(originalContext, storeService, nil, nil, nil, nil, func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		require.Equal(t, originalContext, ctx)
		return wrapperspb.String("module query was called"), nil
	})

	resp, err = QueryModule[wrapperspb.StringValue](accountCtx, &wrapperspb.UInt64Value{Value: 1000})
	require.NoError(t, err)
	require.True(t, proto.Equal(wrapperspb.String("module query was called"), resp))
}
