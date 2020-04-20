package mem_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/mem"
	"github.com/cosmos/cosmos-sdk/store/types"
)

func TestStore(t *testing.T) {
	db := mem.NewStore()
	key, value := []byte("key"), []byte("value")

	require.Equal(t, types.StoreTypeMemory, db.GetStoreType())

	require.Nil(t, db.Get(key))
	db.Set(key, value)
	require.Equal(t, value, db.Get(key))

	newValue := []byte("newValue")
	db.Set(key, newValue)
	require.Equal(t, newValue, db.Get(key))

	db.Delete(key)
	require.Nil(t, db.Get(key))
}

func TestCommit(t *testing.T) {
	db := mem.NewStore()
	key, value := []byte("key"), []byte("value")

	db.Set(key, value)
	id := db.Commit()
	require.True(t, id.IsZero())
	require.True(t, db.LastCommitID().IsZero())
	require.Equal(t, value, db.Get(key))
}
