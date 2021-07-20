package keeper

import (
	"github.com/cosmos/cosmos-sdk/internal/conv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

var (
	ClassKey             = []byte{0x01}
	NFTKey               = []byte{0x02}
	NFTOfClassByOwnerKey = []byte{0x03}
	OwnerKey             = []byte{0x04}
	ClassTotalSupply     = []byte{0x05}

	Delimiter   = []byte{0x00}
	Placeholder = []byte{0x01}
)

// StoreKey is the store key string for nft
const StoreKey = nft.ModuleName

// classStoreKey returns the byte representation of the nft class key
func classStoreKey(classID string) []byte {
	return append(ClassKey, []byte(classID)...)
}

// nftStoreKey returns the byte representation of the nft
func nftStoreKey(classID string) []byte {
	key := append(NFTKey, []byte(classID)...)
	return append(key, Delimiter...)
}

// classTotalSupply returns the byte representation of the ClassTotalSupply
func classTotalSupply(classID string) []byte {
	return append(ClassTotalSupply, []byte(classID)...)
}

// nftOfClassByOwnerStoreKey returns the byte representation of the nft owner
func nftOfClassByOwnerStoreKey(owner sdk.AccAddress, classID string) []byte {
	owner = address.MustLengthPrefix(owner)
	classIDBz := conv.UnsafeStrToBytes(classID)

	var key = make([]byte, 1+len(owner)+1)
	copy(key, NFTOfClassByOwnerKey)
	copy(key[1:], owner)
	copy(key[1+len(owner):], classIDBz)
	return append(key, Delimiter...)
}

// ownerStoreKey returns the byte representation of the nft owner
func ownerStoreKey(classID, nftID string) []byte {
	key := append(OwnerKey, []byte(classID)...)
	key = append(key, Delimiter...)
	return append(key, []byte(nftID)...)
}
