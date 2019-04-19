package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NFT non fungible token definition
type NFT struct {
	Owner       sdk.AccAddress `json:"owner"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Image       string         `json:"image"`
	TokenURI    string         `json:"token_uri"`
}

// NewNFT creates a new NFT
func NewNFT(owner sdk.AccAddress, tokenURI, description, image, name string,
) NFT {
	return NFT{
		Owner:       owner,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Image:       strings.TrimSpace(image),
		TokenURI:    strings.TrimSpace(tokenURI),
	}
}

// EditMetadata edits metadata of an nft
func (nft NFT) EditMetadata(editName, editDescription, editImage, editTokenURI bool,
	name, description, image, tokenURI string) NFT {
	if editName {
		nft.Name = name
	}
	if editDescription {
		nft.Description = description
	}
	if editImage {
		nft.Image = image
	}
	if editTokenURI {
		nft.TokenURI = tokenURI
	}
	return nft
}

func (nft NFT) String() string {
	return fmt.Sprintf(`Owner:              %s
  Name:               %s
  Description:        %s
  Image:        			%s
  TokenURI:   				%s`,
		nft.Owner,
		nft.Name,
		nft.Description,
		nft.Image,
		nft.TokenURI,
	)
}

// Denom is a string
type Denom string

// TokenID is a uint64
type TokenID uint64

// Empty detects whether a TokenID is empty
func (id *TokenID) Empty() bool {
	return id == nil
}

// Owner of non fungible tokens
type Owner map[Denom][]TokenID

// RemoveNFT removes a NFT TokenID from an owner mapping
func (owner Owner) RemoveNFT(denom Denom, id TokenID) (err sdk.Error) {

	// find the index of the NFT as i
	i := 0
	for _, _id := range owner[denom] {
		if _id == id {
			break
		}
		i++
	}

	// NFT Not Found (i will equal len of the array if break was never called)
	if i == len(owner[denom]) {
		return ErrInvalidNFT(DefaultCodespace)
	}

	// remove the ID from the slice
	owner[denom] = append(owner[denom][:i], owner[denom][i+1:]...)
	return
}

// NewOwner returns a new empty owner
func NewOwner() Owner {
	return map[Denom][]TokenID{}
}

// TotalOwnedNFTs gets the total amount of NFTs owned by an account
func (owner Owner) TotalOwnedNFTs() int {
	return len(owner)
}

// Collection of non fungible tokens
type Collection map[TokenID]NFT

// NewCollection creates a new NFT Collection
func NewCollection() Collection {
	return make(map[TokenID]NFT)
}

// GetNFT gets a NFT from the collection
func (collection Collection) GetNFT(denom Denom, id TokenID) (nft NFT, err sdk.Error) {
	nft, ok := collection[id]
	if !ok {
		return nft, ErrUnknownCollection(DefaultCodespace, fmt.Sprintf("collection %s doesn't contain an NFT with TokenID %d", denom, id))
	}
	return
}

// Supply gets the total supply of NFTs of a collection
func (collection Collection) Supply() int {
	return len(collection)
}
