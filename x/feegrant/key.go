package feegrant

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "feegrant"

	// StoreKey is the store key string for supply
	StoreKey = ModuleName

	// RouterKey is the message route for supply
	RouterKey = ModuleName

	// QuerierRoute is the querier route for supply
	QuerierRoute = ModuleName
)

var (
	// FeeAllowanceKeyPrefix is the set of the kvstore for fee allowance data
	// - 0x00<allowance_key_bytes>: allowance
	FeeAllowanceKeyPrefix = []byte{0x00}

	// FeeAllowanceQueueKeyPrefix is the set of the kvstore for fee allowance keys data
	// - 0x01<allowance_prefix_queue_key_bytes>: <empty value>
	FeeAllowanceQueueKeyPrefix = []byte{0x01}
)

// FeeAllowanceKey is the canonical key to store a grant from granter to grantee
// We store by grantee first to allow searching by everyone who granted to you
//
// Key format:
// - <0x00><len(grantee_address_bytes)><grantee_address_bytes><len(granter_address_bytes)><granter_address_bytes>
func FeeAllowanceKey(granter sdk.AccAddress, grantee sdk.AccAddress) []byte {
	return append(FeeAllowancePrefixByGrantee(grantee), address.MustLengthPrefix(granter.Bytes())...)
}

// FeeAllowancePrefixByGrantee returns a prefix to scan for all grants to this given address.
//
// Key format:
// - <0x00><len(grantee_address_bytes)><grantee_address_bytes>
func FeeAllowancePrefixByGrantee(grantee sdk.AccAddress) []byte {
	return append(FeeAllowanceKeyPrefix, address.MustLengthPrefix(grantee.Bytes())...)
}

// FeeAllowancePrefixQueue is the canonical key to store grant key.
//
// Key format:
// - <0x01><exp_bytes><len(grantee_address_bytes)><grantee_address_bytes><len(granter_address_bytes)><granter_address_bytes>
func FeeAllowancePrefixQueue(exp *time.Time, key []byte) []byte {
	allowanceByExpTimeKey := AllowanceByExpTimeKey(exp)
	return append(allowanceByExpTimeKey, key...)
}

// AllowanceByExpTimeKey returns a key with `FeeAllowanceQueueKeyPrefix`, expiry
//
// Key format:
// - <0x01><exp_bytes>
func AllowanceByExpTimeKey(exp *time.Time) []byte {
	// no need of appending len(exp_bytes) here, `FormatTimeBytes` gives const length everytime.
	return append(FeeAllowanceQueueKeyPrefix, sdk.FormatTimeBytes(*exp)...)
}

// ParseAddressesFromFeeAllowanceKey exrtacts and returns the granter, grantee from the given key.
func ParseAddressesFromFeeAllowanceKey(key []byte) (granter, grantee sdk.AccAddress) {
	// key is of format:
	// 0x00<granteeAddressLen (1 Byte)><granteeAddress_Bytes><granterAddressLen (1 Byte)><granterAddress_Bytes>
	kv.AssertKeyAtLeastLength(key, 2)
	granteeAddrLen := key[1] // remove prefix key
	kv.AssertKeyAtLeastLength(key, 2+int(granteeAddrLen))
	grantee = sdk.AccAddress(key[2 : 2+int(granteeAddrLen)])
	granterAddrLen := int(key[2+granteeAddrLen])
	kv.AssertKeyAtLeastLength(key, 3+int(granteeAddrLen)+int(granterAddrLen))
	granter = sdk.AccAddress(key[3+granterAddrLen : 3+int(granteeAddrLen)+int(granterAddrLen)])

	return granter, grantee
}
