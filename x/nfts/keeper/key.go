package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "nfts"

	// StoreKey is the default store key for NFT bank
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the NFT bank store.
	QuerierRoute = StoreKey
)

// keys
var (
	CollectionsKeyPrefix = []byte{0x00} // key for NFT collections
	OwnersNFTsKeyPrefix  = []byte{0x01} // key for balance of NFTs held by an address
)

// TODO: NFTs are stored as follows
// 0x00<denom_bytes><nft_id_bytes><NFT>
// 0x01<address_bytes><denom_bytes><nft_id_bytes><NFT>

// GetCollectionKey gets the key of a collection
func GetCollectionKey(denom string) []byte {
	return append(CollectionsKeyPrefix, []byte(denom)...)
}

// GetOwnersNFTsAddress gets an address from a owners collection key
func GetOwnersNFTsAddress(key []byte) (address sdk.AccAddress) {
	address = key[1:]
	if len(address) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.AccAddress(address)
}

// GetOwnerNFTsKey gets the key of the NFTs owned by an account address
func GetOwnerNFTsKey(address sdk.AccAddress) []byte {
	return append(OwnersNFTsKeyPrefix, address.Bytes()...)
}

// GetOwnerNFTKey gets the key of a NFT owned by an account address
func GetOwnerNFTKey(address sdk.AccAddress, denom string) []byte {
	return append(GetOwnerNFTsKey(address), []byte(denom)...)
}
