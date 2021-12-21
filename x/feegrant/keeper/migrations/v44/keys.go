package v044

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

var (
	GrantPrefix = []byte{0x01}
)

func GrantStoreKey(grantee sdk.AccAddress, granter sdk.AccAddress) []byte {
	granter = address.MustLengthPrefix(granter)
	grantee = address.MustLengthPrefix(grantee)

	l := 1 + len(grantee) + len(granter)
	var key = make([]byte, l)
	copy(key, GrantPrefix)
	copy(key[1:], granter)
	copy(key[1+len(granter):], grantee)

	return key
}

// ParseGrantKey - split granter, grantee address and msg type from the authorization key
func ParseGrantKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress) {
	// key is of format:
	// <granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes>
	kv.AssertKeyAtLeastLength(key, 2)
	granterAddrLen := key[0]
	kv.AssertKeyAtLeastLength(key, int(2+granterAddrLen))
	granterAddr = sdk.AccAddress(key[1 : 1+granterAddrLen])
	granteeAddrLen := int(key[1+granterAddrLen])
	kv.AssertKeyAtLeastLength(key, 3+int(granterAddrLen+byte(granteeAddrLen)))
	granteeAddr = sdk.AccAddress(key[2+granterAddrLen : 2+granterAddrLen+byte(granteeAddrLen)])

	return granterAddr, granteeAddr
}
