package iavl

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/require"
)

func TestImmutableTreePanics(t *testing.T) {
	t.Parallel()
	immTree := iavl.NewImmutableTree(dbm.NewMemDB(), 100, false)
	it := &immutableTree{immTree}
	require.Panics(t, func() { it.Set([]byte{}, []byte{}) })
	require.Panics(t, func() { it.Remove([]byte{}) })
	require.Panics(t, func() { _, _, _ = it.SaveVersion() })
	require.Panics(t, func() { _ = it.DeleteVersion(int64(1)) })

	imm, err := it.GetImmutable(1)
	require.Error(t, err)
	require.Nil(t, imm)

	imm, err = it.GetImmutable(0)
	require.NoError(t, err)
	require.NotNil(t, imm)
	require.Equal(t, immTree, imm)
}
