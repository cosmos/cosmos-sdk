package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Collection of non fungible tokens
type Collection struct {
	Denom string `json:"-"`    // name of the collection; not exported to clients
	NFTs  NFTs   `json:"nfts"` // NFTs that belong to a collection
}

// NewCollection creates a new NFT Collection
func NewCollection(denom string) Collection {
	return Collection{
		Denom: strings.TrimSpace(denom),
		NFTs:  NewNFTs(),
	}
}

// GetNFT gets a NFT from the collection
func (collection Collection) GetNFT(id uint64) (nft NFT, err sdk.Error) {

	for _, nft := range collection.NFTs {
		if nft.ID == id {
			return nft, nil
		}
	}

	return NFT{}, ErrUnknownNFT(DefaultCodespace,
		fmt.Sprintf("NFT #%d doesn't exist on collection %s", nft.ID, collection.Denom),
	)
}

// AddNFT adds an NFT to the collection
func (collection *Collection) AddNFT(nft NFT) {
	collection.NFTs = append(collection.NFTs, nft)
}

// DeleteNFT deletes an NFT from a collection
func (collection *Collection) DeleteNFT(id uint64) {
	// TODO:
}

// Supply gets the total supply of NFTs of a collection
func (collection Collection) Supply() int {
	return len(collection.NFTs)
}
