package keeper

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	PrefixNFT            = []byte{0x01}
	PrefixOwners         = []byte{0x02} // key for a owner
	PrefixCollection     = []byte{0x03} // key for balance of NFTs held by the denom
	PrefixCollectionInfo = []byte{0x04} // key for denom of the nft
	PrefixCollectionName = []byte{0x05} // key for denom name of the nft

	delimiter = []byte("/")
)

// SplitKeyOwner return the address,denom,id from the key of stored owner
func SplitKeyOwner(key []byte) (address sdk.AccAddress, denomID, tokenID string, err error) {
	key = key[len(PrefixOwners)+len(delimiter):]
	keys := bytes.Split(key, delimiter)
	if len(keys) != 3 {
		return address, denomID, tokenID, errors.New("wrong KeyOwner")
	}

	address, _ = sdk.AccAddressFromBech32(string(keys[0]))
	denomID = string(keys[1])
	tokenID = string(keys[2])
	return
}

func SplitKeyDenom(key []byte) (denomID, tokenID string, err error) {
	keys := bytes.Split(key, delimiter)
	if len(keys) != 2 {
		return denomID, tokenID, errors.New("wrong KeyOwner")
	}

	denomID = string(keys[0])
	tokenID = string(keys[1])
	return
}

// KeyOwner gets the key of a collection owned by an account address
func KeyOwner(address sdk.AccAddress, denomID, tokenID string) []byte {
	key := append(PrefixOwners, delimiter...)
	if address != nil {
		key = append(key, []byte(address.String())...)
		key = append(key, delimiter...)
	}

	if address != nil && len(denomID) > 0 {
		key = append(key, []byte(denomID)...)
		key = append(key, delimiter...)
	}

	if address != nil && len(denomID) > 0 && len(tokenID) > 0 {
		key = append(key, []byte(tokenID)...)
	}
	return key
}

// KeyNFT gets the key of nft stored by an denom and id
func KeyNFT(denomID, tokenID string) []byte {
	key := append(PrefixNFT, delimiter...)
	if len(denomID) > 0 {
		key = append(key, []byte(denomID)...)
		key = append(key, delimiter...)
	}

	if len(denomID) > 0 && len(tokenID) > 0 {
		key = append(key, []byte(tokenID)...)
	}
	return key
}

// KeyCollectionSupply gets the storeKey by the collection
func KeyCollectionSupply(denomID string) []byte {
	key := append(PrefixCollection, delimiter...)
	return append(key, []byte(denomID)...)
}

// KeyCollectionInfo gets the storeKey by the denom id
func KeyCollectionInfo(id string) []byte {
	key := append(PrefixCollectionInfo, delimiter...)
	return append(key, []byte(id)...)
}

// KeyCollectionName gets the storeKey by the denom name
func KeyCollectionName(name string) []byte {
	key := append(PrefixCollectionName, delimiter...)
	return append(key, []byte(name)...)
}
