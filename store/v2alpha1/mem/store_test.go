package mem_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/mem"
)

func TestStore(t *testing.T) {
	store := mem.NewStore()
	key, value := []byte("key"), []byte("value")

	require.Equal(t, types.StoreTypeMemory, store.GetStoreType())

	require.Nil(t, store.Get(key))
	store.Set(key, value)
	require.Equal(t, value, store.Get(key))

	newValue := []byte("newValue")
	store.Set(key, newValue)
	require.Equal(t, newValue, store.Get(key))

	store.Delete(key)
	require.Nil(t, store.Get(key))
}

func TestCommit(t *testing.T) {
	store := mem.NewStore()
	key, value := []byte("key"), []byte("value")

	store.Set(key, value)
	id := store.Commit()
	require.True(t, id.IsZero())
	require.True(t, store.LastCommitID().IsZero())
	require.Equal(t, value, store.Get(key))
}
