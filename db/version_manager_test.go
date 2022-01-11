package db_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/db"
)

// Test that VersionManager satisfies the behavior of VersionSet
func TestVersionManager(t *testing.T) {
	vm := db.NewVersionManager(nil)
	require.Equal(t, uint64(0), vm.Last())
	require.Equal(t, 0, vm.Count())
	require.True(t, vm.Equal(vm))
	require.False(t, vm.Exists(0))

	id1, err := vm.Save(0)
	require.NoError(t, err)
	require.Equal(t, uint64(1), id1)
	require.True(t, vm.Exists(id1))
	id2, err := vm.Save(0)
	require.NoError(t, err)
	require.True(t, vm.Exists(id2))
	id3, err := vm.Save(0)
	require.NoError(t, err)
	require.True(t, vm.Exists(id3))

	_, err = vm.Save(id1) // can't save existing id
	require.Error(t, err)

	id4, err := vm.Save(0)
	require.NoError(t, err)
	require.True(t, vm.Exists(id4))
	vm.Delete(id4)
	require.False(t, vm.Exists(id4))

	vm.Delete(id1)
	require.False(t, vm.Exists(id1))
	require.Equal(t, id2, vm.Initial())
	require.Equal(t, id3, vm.Last())

	var all []uint64
	for it := vm.Iterator(); it.Next(); {
		all = append(all, it.Value())
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	require.Equal(t, []uint64{id2, id3}, all)

	vmc := vm.Copy()
	id5, err := vmc.Save(0)
	require.NoError(t, err)
	require.False(t, vm.Exists(id5)) // true copy is made

	vm2 := db.NewVersionManager([]uint64{id2, id3})
	require.True(t, vm.Equal(vm2))
}
