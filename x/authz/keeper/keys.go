package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/internal/conv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// Keys for store prefixes
// Items are stored with the following key: values
//
// - 0x01<grant_Bytes>: Grant
// - 0x02<grant_expiration_Bytes>: GrantQueueItem
//
var (
	GrantKey         = []byte{0x01} // prefix for each key
	GrantQueuePrefix = []byte{0x02}
)

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

// StoreKey is the store key string for authz
const StoreKey = authz.ModuleName

// grantStoreKey - return authorization store key
// Items are stored with the following key: values
//
// - 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>: Grant
func grantStoreKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
	m := conv.UnsafeStrToBytes(msgType)
	granter = address.MustLengthPrefix(granter)
	grantee = address.MustLengthPrefix(grantee)

	l := 1 + len(grantee) + len(granter) + len(m)
	var key = make([]byte, l)
	copy(key, GrantKey)
	copy(key[1:], granter)
	copy(key[1+len(granter):], grantee)
	copy(key[l-len(m):], m)
	//	fmt.Println(">>>> len", l, key)
	return key
}

// parseGrantStoreKey - split granter, grantee address and msg type from the authorization key
func parseGrantStoreKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress, msgType string) {
	// key is of format:
	// 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>
	kv.AssertKeyAtLeastLength(key, 2)
	granterAddrLen := key[1] // remove prefix key
	kv.AssertKeyAtLeastLength(key, int(3+granterAddrLen))
	granterAddr = sdk.AccAddress(key[2 : 2+granterAddrLen])
	granteeAddrLen := int(key[2+granterAddrLen])
	kv.AssertKeyAtLeastLength(key, 4+int(granterAddrLen+byte(granteeAddrLen)))
	granteeAddr = sdk.AccAddress(key[3+granterAddrLen : 3+granterAddrLen+byte(granteeAddrLen)])

	return granterAddr, granteeAddr, conv.UnsafeBytesToStr(key[3+granterAddrLen+byte(granteeAddrLen):])
}

// parseGrantQueueKey split expiration time, granter and grantee from the grant queue key
func parseGrantQueueKey(key []byte) (time.Time, sdk.AccAddress, sdk.AccAddress, error) {
	// key is of format:
	// 0x02<grant_expiration_Bytes><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes>

	kv.AssertKeyAtLeastLength(key, 1+lenTime)
	exp, err := sdk.ParseTimeBytes(key[1 : 1+lenTime])
	if err != nil {
		return exp, nil, nil, err
	}

	granterAddrLen := key[1+lenTime]
	kv.AssertKeyAtLeastLength(key, 1+lenTime+int(granterAddrLen))
	granter := sdk.AccAddress(key[2+lenTime : byte(2+lenTime)+granterAddrLen])

	granteeAddrLen := key[byte(2+lenTime)+granterAddrLen]
	granteeStart := byte(3+lenTime) + granterAddrLen
	kv.AssertKeyAtLeastLength(key, int(granteeStart))
	grantee := sdk.AccAddress(key[granteeStart : granteeStart+granteeAddrLen])

	return exp, granter, grantee, nil
}

// GrantQueueKey - return grant queue store key
// Key format is
//
// - 0x02<grant_expiration_Bytes>: GrantQueueItem
func GrantQueueKey(expiration time.Time, granter sdk.AccAddress, grantee sdk.AccAddress) []byte {
	exp := sdk.FormatTimeBytes(expiration)
	granter = address.MustLengthPrefix(granter)
	grantee = address.MustLengthPrefix(grantee)

	l := 1 + len(exp) + len(granter) + len(grantee)
	var key = make([]byte, l)
	copy(key, GrantQueuePrefix)
	copy(key[1:], exp)
	copy(key[1+len(exp):], granter)
	copy(key[1+len(exp)+len(granter):], grantee)
	return key
}

// GrantQueueTimePrefix - return grant queue time prefix
func GrantQueueTimePrefix(expiration time.Time) []byte {
	return append(GrantQueuePrefix, sdk.FormatTimeBytes(expiration)...)
}

// firstAddressFromGrantStoreKey parses the first address only
func firstAddressFromGrantStoreKey(key []byte) sdk.AccAddress {
	addrLen := key[0]
	return sdk.AccAddress(key[1 : 1+addrLen])
}
