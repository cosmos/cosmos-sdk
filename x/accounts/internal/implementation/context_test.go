package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
)

func TestMakeAccountContext(t *testing.T) {
	storeService, ctx := colltest.MockStore()
	accountAddr := []byte("accountAddr")
	sender := []byte("sender")
	sb := collections.NewSchemaBuilderFromAccessor(OpenKVStore)

	accountCtx := MakeAccountContext(ctx, storeService, accountAddr, sender)

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
	store := storeService.OpenKVStore(ctx)

	// now we want the value to be store in the following accounts prefix (accountAddr + itemPrefix)
	value, err := store.Get(append(accountAddr, itemPrefix...))
	require.NoError(t, err)
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 3, 232}, value)
}
