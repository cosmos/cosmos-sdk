package address_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/address"
)

func TestLengthPrefixedAddressStoreKey(t *testing.T) {
	addr10byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	addr256byte := make([]byte, 256)

	tests := []struct {
		name        string
		addr        []byte
		expStoreKey []byte
		expErr      bool
	}{
		{"10-byte address", addr10byte, append([]byte{byte(10)}, addr10byte...), false},
		{"20-byte address", addr20byte, append([]byte{byte(20)}, addr20byte...), false},
		{"256-byte address (too long)", addr256byte, nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			storeKey, err := address.LengthPrefixedStoreKey(tt.addr)
			if tt.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expStoreKey, storeKey)
			}
		})
	}
}
