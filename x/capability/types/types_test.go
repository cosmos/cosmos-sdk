package types_test

import (
	"fmt"
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
	o1 := types.NewOwner("bank", "send")
	o2 := types.NewOwner("slashing", "slash")

	require.NoError(t, co.Set(o1))
	require.Error(t, co.Set(o1))
	require.NoError(t, co.Set(o2))
	require.Equal(t, []types.Owner{o1, o2}, co.Owners)
}
