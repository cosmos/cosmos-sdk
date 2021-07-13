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
)

// StoreKey is the store key string for nft
const StoreKey = nft.ModuleName

// classStoreKey returns the byte representation of the nft class key
func classStoreKey(classID string) []byte {
	return append(ClassKey, []byte(classID)...)
}

// nftStoreKey returns the byte representation of the nft
func nftStoreKey(classID string) []byte {
	return append(NFTKey, []byte(classID)...)
}

// classTotalSupply returns the byte representation of the ClassTotalSupply
func classTotalSupply(classID string) []byte {
	return append(ClassTotalSupply, []byte(classID)...)
}

// nftOfClassByOwnerStoreKey returns the byte representation of the nft owner
func nftOfClassByOwnerStoreKey(owner sdk.AccAddress, classID string) []byte {
	owner = address.MustLengthPrefix(owner)
	m := conv.UnsafeStrToBytes(classID)

	var key = make([]byte, 1+len(owner))
	copy(key, NFTOfClassByOwnerKey)
	copy(key[1:], owner)
	copy(key[1+len(owner):], m)
	return key
}

// ownerStoreKey returns the byte representation of the nft owner
func ownerStoreKey(classID, nftID string) []byte {
	key := append(OwnerKey, []byte(classID)...)
	return append(key, []byte(nftID)...)
}
