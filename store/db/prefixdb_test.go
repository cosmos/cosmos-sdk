package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/store/db"
	"cosmossdk.io/store/mock"
)

func TestPrefixDB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock.NewMockKVStoreWithBatch(mockCtrl)
	prefix := []byte("test:")
	pdb := db.NewPrefixDB(mockDB, prefix)

	key := []byte("key1")
	value := []byte("value1")
	mockDB.EXPECT().Set(gomock.Eq(append(prefix, key...)), gomock.Eq(value)).Return(nil)

	err := pdb.Set(key, value)
	require.NoError(t, err)

	mockDB.EXPECT().Get(gomock.Eq(append(prefix, key...))).Return(value, nil)

	returnedValue, err := pdb.Get(key)
	require.NoError(t, err)
	require.Equal(t, value, returnedValue)

	mockDB.EXPECT().Has(gomock.Eq(append(prefix, key...))).Return(true, nil)

	has, err := pdb.Has(key)
	require.NoError(t, err)
	require.True(t, has)

	mockDB.EXPECT().Delete(gomock.Eq(append(prefix, key...))).Return(nil)

	err = pdb.Delete(key)
	require.NoError(t, err)

	mockDB.EXPECT().Has(gomock.Eq(append(prefix, key...))).Return(false, nil)

	has, err = pdb.Has(key)
	require.NoError(t, err)
	require.False(t, has)
}
