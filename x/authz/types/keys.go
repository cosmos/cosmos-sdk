package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
// - 0x01<granterAddress_Bytes><granteeAddress_Bytes><msgType_Bytes>: Grant

var (
	// Keys for store prefixes
	GrantKey = []byte{0x01} // prefix for each key
)

// GetActorAuthorizationKey - return authorization store key
func GetActorAuthorizationKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
	return append(append(append(GrantKey, granter.Bytes()...), grantee.Bytes()...), []byte(msgType)...)

}

// ExtractAddressesFromGrantKey - split granter & grantee address from the authorization key
func ExtractAddressesFromGrantKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress) {
	granterAddr = sdk.AccAddress(key[1 : sdk.AddrLen+1])
	granteeAddr = sdk.AccAddress(key[sdk.AddrLen+1 : sdk.AddrLen*2+1])
	return granterAddr, granteeAddr
}
