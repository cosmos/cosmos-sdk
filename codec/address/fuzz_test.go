package address

import (
	"cosmossdk.io/core/address"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	sdkAddress "github.com/cosmos/cosmos-sdk/types/address"
)

func FuzzCachedAddressCodec(f *testing.F) {
	if testing.Short() {
		f.Skip()
	}

	addresses := generateAddresses(2)
	for _, addr := range addresses {
		f.Add(addr)
	}
	cdc := NewBech32Codec("cosmos")

	f.Fuzz(func(t *testing.T, addr []byte) {
		checkAddress(t, addr, cdc)
	})

}

func FuzzAddressCodec(f *testing.F) {
	if testing.Short() {
		f.Skip()
	}
	addresses := generateAddresses(2)
	for _, addr := range addresses {
		f.Add(addr)
	}

	cdc := Bech32Codec{Bech32Prefix: "cosmos"}

	f.Fuzz(func(t *testing.T, addr []byte) {
		checkAddress(t, addr, cdc)
	})

}

func checkAddress(t *testing.T, addr []byte, cdc address.Codec) {
	t.Helper()
	if len(addr) > sdkAddress.MaxAddrLen {
		return
	}
	strAddr, err := cdc.BytesToString(addr)
	if err != nil {
		t.Fatal(err)
	}
	b, err := cdc.StringToBytes(strAddr)
	if err != nil {
		if !errors.Is(errEmptyAddress, err) {
			t.Fatal(err)
		}
	}
	require.Equal(t, len(addr), len(b))
}
