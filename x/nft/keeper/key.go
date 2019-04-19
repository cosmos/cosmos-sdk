package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "nft"

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
func GetCollectionKey(denom types.Denom) []byte {
	return append(collectionKeyPrefix, []byte(denom)...)
}

// GetOwnerKey gets the key of a collection
func GetOwnerKey(address sdk.AccAddress) []byte {
	return append(ownerKeyPrefix, address.Bytes()...)
}
