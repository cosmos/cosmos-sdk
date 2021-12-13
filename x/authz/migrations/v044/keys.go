package v044

import (
	"time"

	"github.com/cosmos/cosmos-sdk/internal/conv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// Keys for store prefixes
// Items are stored with the following key: values
//
// - 0x01<grant_Bytes>: Grant
//
var (
	GrantKey         = []byte{0x01}
	GrantQueuePrefix = []byte{0x02}
)

// GrantStoreKey - return authorization store key
// Items are stored with the following key: values
//
// - 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>: Grant
func GrantStoreKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
	m := conv.UnsafeStrToBytes(msgType)
	granter = address.MustLengthPrefix(granter)
	grantee = address.MustLengthPrefix(grantee)

	l := 1 + len(grantee) + len(granter) + len(m)
	var key = make([]byte, l)
	copy(key, GrantKey)
	copy(key[1:], granter)
	copy(key[1+len(granter):], grantee)
	copy(key[l-len(m):], m)

	return key
}

// ParseGrantStoreKey - split granter, grantee address and msg type from the authorization key
func ParseGrantStoreKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress, msgType string) {
	// key is of format:
	// <granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>
	kv.AssertKeyAtLeastLength(key, 2)
	granterAddrLen := key[0] 
	kv.AssertKeyAtLeastLength(key, int(3+granterAddrLen))
	granterAddr = sdk.AccAddress(key[2 : 2+granterAddrLen])
	granteeAddrLen := int(key[2+granterAddrLen])
	kv.AssertKeyAtLeastLength(key, 4+int(granterAddrLen+byte(granteeAddrLen)))
	granteeAddr = sdk.AccAddress(key[3+granterAddrLen : 3+granterAddrLen+byte(granteeAddrLen)])

	return granterAddr, granteeAddr, conv.UnsafeBytesToStr(key[3+granterAddrLen+byte(granteeAddrLen):])
}

// GrantQueueKey - return grant queue store key
// Key format is
//
// - 0x02<grant_expiration_Bytes>:
func GrantQueueKey(expiration time.Time) []byte {
	return append(GrantQueuePrefix, sdk.FormatTimeBytes(expiration)...)
}
