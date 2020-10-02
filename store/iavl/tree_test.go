package iavl

import (
	"testing"

	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestImmutableTreePanics(t *testing.T) {
	t.Parallel()
	immTree := iavl.NewImmutableTree(dbm.NewMemDB(), 100)
	it := &immutableTree{immTree}
	require.Panics(t, func() { it.Set([]byte{}, []byte{}) })
	require.Panics(t, func() { it.Remove([]byte{}) })
	require.Panics(t, func() { it.SaveVersion() })           // nolint:errcheck
	require.Panics(t, func() { it.DeleteVersion(int64(1)) }) // nolint:errcheck
	v, _ := it.GetVersioned([]byte{0x01}, 1)
	require.Equal(t, int64(-1), v)
	v, _ = it.GetVersioned([]byte{0x01}, 0)
	require.Equal(t, int64(0), v)

	val, proof, err := it.GetVersionedWithProof(nil, 1)
	require.Error(t, err)
	require.Nil(t, val)
	require.Nil(t, proof)

	imm, err := it.GetImmutable(1)
	require.Error(t, err)
	require.Nil(t, imm)

	imm, err = it.GetImmutable(0)
	require.NoError(t, err)
	require.NotNil(t, imm)
	require.Equal(t, immTree, imm)
}
