package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NFTs are stored as follow:
//
// - Colections: 0x00<denom_bytes_key> :<Collection>
//
// - Balances: 0x01<address_bytes_key><denom_bytes_key>: <Collections>
var (
	CollectionsKeyPrefix = []byte{0x00} // key for NFT collections
	BalancesKeyPrefix    = []byte{0x01} // key for balance of NFTs held by an address 
)

// GetCollectionKey gets the key of a collection
func GetCollectionKey(denom string) []byte {
	// TODO: use the hash of the denom instead
	return append(CollectionsKeyPrefix, []byte(denom)...)
}

// SplitBalanceKey gets an address and denom from a balance key
func SplitBalanceKey(key []byte) (sdk.AccAddress, string) {
	if len(key) != sdk.AddrLen {
		panic("unexpected key length")
	}
	address := key[1 : sdk.AddrLen+1]
	denomBz := key[sdk.AddrLen+1:] // TODO: use the hash as the key
	return sdk.AccAddress(address), string(denomBz)
}

// GetBalancesKey gets the key prefix for all the collections owned by an account address
func GetBalancesKey(address sdk.AccAddress) []byte {
	return append(BalancesKeyPrefix, address.Bytes()...)
}

// GetBalanceKey gets the key of a collection owned by an account address
func GetBalanceKey(address sdk.AccAddress, denom string) []byte {
	// TODO: use the hash of the denom instead
	return append(GetBalancesKey(address), []byte(denom)...)
}
