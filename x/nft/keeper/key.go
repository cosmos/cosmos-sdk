package keeper

import (
	"crypto/sha1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NFTs are stored as follow:
//
// - Colections: 0x00<denom_bytes_key> :<Collection>
//
// - Balances: 0x01<address_bytes_key><denom_bytes_key>: <Collections>
var (
	CollectionsKeyPrefix = []byte{0x00} // key for NFT collections
	OwnersKeyPrefix      = []byte{0x01} // key for balance of NFTs held by an address
)

// GetCollectionKey gets the key of a collection
func GetCollectionKey(denom string) []byte {

	h := sha1.New()
	_, err := h.Write([]byte(denom))
	if err != nil {
		panic(err)
	}
	bs := h.Sum(nil)

	return append(CollectionsKeyPrefix, bs...)
}

// SplitOwnerKey gets an address and denom from an owner key
func SplitOwnerKey(key []byte) (sdk.AccAddress, []byte) {
	if len(key) != sdk.AddrLen {
		panic("unexpected key length")
	}
	address := key[1 : sdk.AddrLen+1]
	denomHashBz := key[sdk.AddrLen+1:]
	return sdk.AccAddress(address), denomHashBz
}

// GetOwnersKey gets the key prefix for all the collections owned by an account address
func GetOwnersKey(address sdk.AccAddress) []byte {
	return append(OwnersKeyPrefix, address.Bytes()...)
}

// GetOwnerKey gets the key of a collection owned by an account address
func GetOwnerKey(address sdk.AccAddress, denom string) []byte {

	h := sha1.New()
	_, err := h.Write([]byte(denom))
	if err != nil {
		panic(err)
	}
	bs := h.Sum(nil)

	return append(GetOwnersKey(address), bs...)
}
