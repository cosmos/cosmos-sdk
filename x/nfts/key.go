package nfts

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName is the name of the module
	ModuleName = "nfts"

	// StoreKey is the default store key for NFT bank
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the NFT bank store.
	QuerierRoute = StoreKey
)

var (
	collectionKeyPrefix = []byte{0x00}
	ownerKeyPrefix      = []byte{0x01}
)

// GetCollectionKey gets the key of a collection
func GetCollectionKey(denom Denom) []byte {
	return append(collectionKeyPrefix, []byte(denom)...)
}

// GetOwnerKey gets the key of a collection
func GetOwnerKey(address sdk.AccAddress) []byte {
	return append(ownerKeyPrefix, address.Bytes()...)
}
