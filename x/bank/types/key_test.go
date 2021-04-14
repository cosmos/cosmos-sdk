package types_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func cloneAppend(bz []byte, tail []byte) (res []byte) {
	res = make([]byte, len(bz)+len(tail))
	copy(res, bz)
	copy(res[len(bz):], tail)
	return
}

func TestAddressFromBalancesStore(t *testing.T) {
	addr, err := sdk.AccAddressFromBech32("cosmos1n88uc38xhjgxzw9nwre4ep2c8ga4fjxcar6mn7")
	require.NoError(t, err)
	addrLen := len(addr)
	require.Equal(t, 20, addrLen)

	key := cloneAppend(address.MustLengthPrefix(addr), []byte("stake"))
	res, err := types.AddressFromBalancesStore(key)
	require.NoError(t, err)
	require.Equal(t, res, addr)
}

func TestInvalidAddressFromBalancesStore(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{"empty", []byte("")},
		{"invalid", []byte("3AA")},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := types.AddressFromBalancesStore(tc.key)
			assert.Error(t, err)
			assert.True(t, errors.Is(types.ErrInvalidKey, err))
		})
	}
}
