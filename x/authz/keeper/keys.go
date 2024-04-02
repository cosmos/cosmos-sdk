package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/x/authz"
	"cosmossdk.io/x/authz/internal/conv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// Keys for store prefixes
// Items are stored with the following key: values
//
// - 0x01<grant_Bytes>: Grant
// - 0x02<grant_expiration_Bytes>: GrantQueueItem
var (
	GrantKey         = []byte{0x01} // prefix for each key
	GrantQueuePrefix = []byte{0x02}
)

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

const (
	// StoreKey is the store key string for authz
	StoreKey = authz.ModuleName

	bytePositionOfGrantKeyPrefix    = 0
	bytePositionOfGranterAddressLen = 1
	omitTwoBytes                    = 2
)

// grantStoreKey - return authorization store key
// Items are stored with the following key: values
//
// - 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>: Grant
func grantStoreKey(grantee, granter sdk.AccAddress, msgType string) []byte {
	m := conv.UnsafeStrToBytes(msgType)
	granter = address.MustLengthPrefix(granter)
	grantee = address.MustLengthPrefix(grantee)
	key := sdk.AppendLengthPrefixedBytes(GrantKey, granter, grantee, m)

	return key
}

// parseGrantStoreKey - split granter, grantee address and msg type from the authorization key
func parseGrantStoreKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress, msgType string) {
	// key is of format:
	// 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>

	granterAddrLen, granterAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1) // ignore key[0] since it is a prefix key
	granterAddr, granterAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, granterAddrLenEndIndex+1, int(granterAddrLen[0]))

	granteeAddrLen, granteeAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, granterAddrEndIndex+1, 1)
	granteeAddr, granteeAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, granteeAddrLenEndIndex+1, int(granteeAddrLen[0]))

	kv.AssertKeyAtLeastLength(key, granteeAddrEndIndex+1)
	return granterAddr, granteeAddr, conv.UnsafeBytesToStr(key[(granteeAddrEndIndex + 1):])
}

// parseGrantQueueKey split expiration time, granter and grantee from the grant queue key
func parseGrantQueueKey(key []byte) (time.Time, sdk.AccAddress, sdk.AccAddress, error) {
	// key is of format:
	// 0x02<grant_expiration_Bytes><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes>

	expBytes, expEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, lenTime)

	exp, err := sdk.ParseTimeBytes(expBytes)
	if err != nil {
		return exp, nil, nil, err
	}

	granterAddrLen, granterAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, expEndIndex+1, 1)
	granter, granterEndIndex := sdk.ParseLengthPrefixedBytes(key, granterAddrLenEndIndex+1, int(granterAddrLen[0]))

	granteeAddrLen, granteeAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, granterEndIndex+1, 1)
	grantee, _ := sdk.ParseLengthPrefixedBytes(key, granteeAddrLenEndIndex+1, int(granteeAddrLen[0]))

	return exp, granter, grantee, nil
}

// GrantQueueKey - return grant queue store key. If a given grant doesn't have a defined
// expiration, then it should not be used in the pruning queue.
// Key format is:
//
//	0x02<expiration><granterAddressLen (1 Byte)><granterAddressBytes><granteeAddressLen (1 Byte)><granteeAddressBytes>: GrantQueueItem
func GrantQueueKey(expiration time.Time, granter, grantee sdk.AccAddress) []byte {
	exp := sdk.FormatTimeBytes(expiration)
	granter = address.MustLengthPrefix(granter)
	grantee = address.MustLengthPrefix(grantee)

	return sdk.AppendLengthPrefixedBytes(GrantQueuePrefix, exp, granter, grantee)
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

// grantKeyToString converts a byte slice representing a grant key into a human-readable UTF-8 encoded string.
// The expected format of the byte slice is as follows:
// 0x01<prefix(1 Byte)><granterAddressLen(1 Byte)><granterAddress_Bytes><granteeAddressLen(1 Byte)><granteeAddress_Bytes><msgType_Bytes>
func grantKeyToString(skey []byte) string {
	// get grant key prefix
	prefix := skey[bytePositionOfGrantKeyPrefix]

	// get granter address len and granter address, found at a specified position in the byte slice
	granterAddressLen := int(skey[bytePositionOfGranterAddressLen])
	startByteIndex := omitTwoBytes
	endByteIndex := startByteIndex + granterAddressLen
	granterAddressBytes := skey[startByteIndex:endByteIndex]

	// get grantee address len and grantee address, found at a specified position in the byte slice
	granteeAddressLen := int(skey[endByteIndex])
	startByteIndex = endByteIndex + 1
	endByteIndex = startByteIndex + granteeAddressLen
	granteeAddressBytes := skey[startByteIndex:endByteIndex]

	// get message type, start from the specific byte to the end
	startByteIndex = endByteIndex
	msgTypeBytes := skey[startByteIndex:]

	// build UTF-8 encoded string
	granterAddr := sdk.AccAddress(granterAddressBytes)
	granteeAddr := sdk.AccAddress(granteeAddressBytes)
	return fmt.Sprintf("%d|%d|%s|%d|%s|%s", prefix, granterAddressLen, granterAddr.String(), granteeAddressLen, granteeAddr.String(), msgTypeBytes)
}
