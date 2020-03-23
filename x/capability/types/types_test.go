package types_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

func TestCapabilityKey(t *testing.T) {
	idx := uint64(3162)
	cap := types.NewCapabilityKey(idx)
	require.Equal(t, idx, cap.GetIndex())
	require.Equal(t, fmt.Sprintf("CapabilityKey{%p, %d}", cap, idx), cap.String())
}

func TestOwner(t *testing.T) {
	o := types.NewOwner("bank", "send")
	require.Equal(t, "bank/send", o.Key())
	require.Equal(t, "module: bank\nname: send\n", o.String())
}

func TestCapabilityOwners(t *testing.T) {
	co := types.NewCapabilityOwners()

	owners := make([]types.Owner, 1024)
	for i := range owners {
		var owner types.Owner

		if i%2 == 0 {
			owner = types.NewOwner("bank", fmt.Sprintf("send-%d", i))
		} else {
			owner = types.NewOwner("slashing", fmt.Sprintf("slash-%d", i))
		}

		owners[i] = owner
		require.NoError(t, co.Set(owner))
	}

	sort.Slice(owners, func(i, j int) bool { return owners[i].Key() < owners[j].Key() })
	require.Equal(t, owners, co.Owners)

	for _, owner := range owners {
		require.Error(t, co.Set(owner))
	}
}
