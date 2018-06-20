package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dbm "github.com/tendermint/tmlibs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestStore(t *testing.T) {
	db := dbm.NewMemDB()
	cms := NewCommitMultiStore(db)

	key := sdk.NewKVStoreKey("test")
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	assert.Nil(t, err)

	store := cms.GetKVStore(key)
	assert.NotNil(t, store)

	k := []byte("hello")
	v := []byte("world")
	assert.False(t, store.Has(k))
	store.Set(k, v)
	assert.True(t, store.Has(k))
	assert.Equal(t, v, store.Get(k))
	store.Delete(k)
	assert.False(t, store.Has(k))
}
