package types

import (
	"testing"

	"cosmossdk.io/collections/colltest"
)

func TestCollectionsCorrectness(t *testing.T) {
	t.Run("AccAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, AccAddressKey, AccAddress{0x0, 0x2, 0x3, 0x5})
	})

	t.Run("ValAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, ValAddressKey, ValAddress{0x1, 0x3, 0x4})
	})

	t.Run("ConsAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, ConsAddressKey, ConsAddress{0x32, 0x0, 0x0, 0x3})
	})

	t.Run("AddressIndexingKey", func(t *testing.T) {
		colltest.TestKeyCodec(t, AddressKeyAsIndexKey(AccAddressKey), AccAddress{0x2, 0x5, 0x8})
	})
}
