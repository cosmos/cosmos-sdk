package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "authz"

	// StoreKey is the store key string for authz
	StoreKey = ModuleName

	// RouterKey is the message route for authz
	RouterKey = ModuleName

	// QuerierRoute is the querier route for authz
	QuerierRoute = ModuleName
)

// Keys for authz store
// Items are stored with the following key: values
//
// - 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>: Grant

var (
	// Keys for store prefixes
	GrantKey = []byte{0x01} // prefix for each key
)

// GetAuthorizationStoreKey - return authorization store key
func GetAuthorizationStoreKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
	return append(append(append(
		GrantKey,
		address.MustLengthPrefix(granter)...),
		address.MustLengthPrefix(grantee)...),
		[]byte(msgType)...,
	)
}

// ExtractAddressesFromGrantKey - split granter & grantee address from the authorization key
func ExtractAddressesFromGrantKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress) {
	// key if of format:
	// 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>
	granterAddrLen := key[1] // remove prefix key
	granterAddr = sdk.AccAddress(key[2 : 2+granterAddrLen])
	granteeAddrLen := int(key[2+granterAddrLen])
	granteeAddr = sdk.AccAddress(key[3+granterAddrLen : 3+granterAddrLen+byte(granteeAddrLen)])

	return granterAddr, granteeAddr
}
