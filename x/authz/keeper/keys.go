package keeper

import (
	"github.com/cosmos/cosmos-sdk/internal/conv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// Keys for store prefixes
var (
	GrantKey = []byte{0x01} // prefix for each key
)

// StoreKey is the store key string for authz
const StoreKey = authz.ModuleName

// grantStoreKey - return authorization store key
// Items are stored with the following key: values
//
// - 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>: Grant
func grantStoreKey(grantee, granter sdk.AccAddress, msgType string) []byte {
	m := conv.UnsafeStrToBytes(msgType)
	granter = address.MustLengthPrefix(granter)
	grantee = address.MustLengthPrefix(grantee)

	l := 1 + len(grantee) + len(granter) + len(m)
	key := make([]byte, l)
	copy(key, GrantKey)
	copy(key[1:], granter)
	copy(key[1+len(granter):], grantee)
	copy(key[l-len(m):], m)
	//	fmt.Println(">>>> len", l, key)
	return key
}

// addressesFromGrantStoreKey - split granter & grantee address from the authorization key
func addressesFromGrantStoreKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress) {
	// key is of format:
	// 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>
	kv.AssertKeyAtLeastLength(key, 2)
	granterAddrLen := key[1] // remove prefix key
	kv.AssertKeyAtLeastLength(key, int(3+granterAddrLen))
	granterAddr = sdk.AccAddress(key[2 : 2+granterAddrLen])
	granteeAddrLen := int(key[2+granterAddrLen])
	kv.AssertKeyAtLeastLength(key, 4+int(granterAddrLen+byte(granteeAddrLen)))
	granteeAddr = sdk.AccAddress(key[3+granterAddrLen : 3+granterAddrLen+byte(granteeAddrLen)])

	return granterAddr, granteeAddr
}

// firstAddressFromGrantStoreKey parses the first address only
func firstAddressFromGrantStoreKey(key []byte) sdk.AccAddress {
	addrLen := key[0]
	return sdk.AccAddress(key[1 : 1+addrLen])
}
