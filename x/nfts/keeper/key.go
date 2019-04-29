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

// NFTs are stored as follow:
//
// - Colections: 0x00<denom_bytes_key><Collection>
//
// - Balances: 0x01<address_bytes_key><denom_bytes_key><Collection>
var (
	CollectionsKeyPrefix = []byte{0x00} // key for NFT collections
	NFTBalancesKeyPrefix = []byte{0x01} // key for balance of NFTs held by an address
)

// GetCollectionKey gets the key of a collection
func GetCollectionKey(denom string) []byte {
	return append(CollectionsKeyPrefix, []byte(denom)...)
}

// GetNFTBalancesAddress gets an address from an account NFT balance key
func GetNFTBalancesAddress(key []byte) (address sdk.AccAddress) {
	address = key[1:]
	if len(address) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.AccAddress(address)
}

// GetNFTBalancesKey gets the key of the NFTs owned by an account address
func GetNFTBalancesKey(address sdk.AccAddress) []byte {
	return append(NFTBalancesKeyPrefix, address.Bytes()...)
}

// GetBalancesNFTKey gets the key of a single NFT owned by an account address
func GetBalancesNFTKey(address sdk.AccAddress, denom string) []byte {
	return append(GetNFTBalancesKey(address), []byte(denom)...)
}
