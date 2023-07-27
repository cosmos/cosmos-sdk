package feegrant

import (
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

// FeeAllowanceKeyPrefix is the set of the kvstore for fee allowance data
var FeeAllowanceKeyPrefix = []byte{0x00}

// FeeAllowanceKey is the canonical key to store a grant from granter to grantee
// We store by grantee first to allow searching by everyone who granted to you
func FeeAllowanceKey(granter, grantee sdk.AccAddress) []byte {
	return append(FeeAllowancePrefixByGrantee(grantee), address.MustLengthPrefix(granter.Bytes())...)
}

// FeeAllowancePrefixByGrantee returns a prefix to scan for all grants to this given address.
func FeeAllowancePrefixByGrantee(grantee sdk.AccAddress) []byte {
	return append(FeeAllowanceKeyPrefix, address.MustLengthPrefix(grantee.Bytes())...)
}

func ParseAddressesFromFeeAllowanceKey(key []byte) (granter, grantee sdk.AccAddress) {
	// key is of format:
	// 0x00<granteeAddressLen (1 Byte)><granteeAddress_Bytes><granterAddressLen (1 Byte)><granterAddress_Bytes><msgType_Bytes>
	kv.AssertKeyAtLeastLength(key, 2)
	granteeAddrLen := key[1] // remove prefix key
	kv.AssertKeyAtLeastLength(key, int(2+granteeAddrLen))
	grantee = sdk.AccAddress(key[2 : 2+granteeAddrLen])
	granterAddrLen := int(key[2+granteeAddrLen])
	kv.AssertKeyAtLeastLength(key, 3+int(granteeAddrLen+byte(granterAddrLen)))
	granter = sdk.AccAddress(key[3+granterAddrLen : 3+granteeAddrLen+byte(granterAddrLen)])

	return granter, grantee
}
