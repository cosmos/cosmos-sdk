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

// parseGrantQueueKey split expiration time from the grant queue key
func parseGrantQueueKey(key []byte) (expiration time.Time) {
	// key is of format:
	// 0x02<grant_expiration_Bytes>

	expiration, err := sdk.ParseTimeBytes(key[1 : 1+lenTime])
	if err != nil {
		panic(err)
	}

	return expiration
}

// GrantQueueKey - return grant queue store key
// Key format is
//
// - 0x02<grant_expiration_Bytes>: GrantQueueItem
func GrantQueueKey(expiration time.Time) []byte {
	return append(GrantQueuePrefix, sdk.FormatTimeBytes(expiration)...)
}
