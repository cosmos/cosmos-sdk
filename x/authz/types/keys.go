package types

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
