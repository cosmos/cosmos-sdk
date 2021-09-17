package db_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-sdk/db"
)

// Test that VersionManager satisfies the behavior of VersionSet
func TestVersionManager(t *testing.T) {
	vm := dbm.NewVersionManager(nil)
	require.Equal(t, uint64(0), vm.Last())
	require.Equal(t, 0, vm.Count())
	require.True(t, vm.Equal(vm))
	require.False(t, vm.Exists(0))

	id, err := vm.Save(0)
	require.NoError(t, err)
	require.Equal(t, uint64(1), id)
	require.True(t, vm.Exists(id))
	id2, err := vm.Save(0)
	require.NoError(t, err)
	require.True(t, vm.Exists(id2))
	id3, err := vm.Save(0)
	require.NoError(t, err)
	require.True(t, vm.Exists(id3))

	id, err = vm.Save(id) // can't save existing id
	require.Error(t, err)

	id, err = vm.Save(0)
	require.NoError(t, err)
	require.True(t, vm.Exists(id))
	vm.Delete(id)
	require.False(t, vm.Exists(id))

	vm.Delete(1)
	require.False(t, vm.Exists(1))
	require.Equal(t, id2, vm.Initial())
	require.Equal(t, id3, vm.Last())

	var all []uint64
	for it := vm.Iterator(); it.Next(); {
		all = append(all, it.Value())
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	require.Equal(t, []uint64{id2, id3}, all)

	vmc := vm.Copy()
	id, err = vmc.Save(0)
	require.NoError(t, err)
	require.False(t, vm.Exists(id)) // true copy is made

	vm2 := dbm.NewVersionManager([]uint64{id2, id3})
	require.True(t, vm.Equal(vm2))
}
