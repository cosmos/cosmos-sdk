package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: create interface for buyable NFT

// NFT non fungible token interface
type NFT interface {
	EditMetadata(editName, editDescription, editImage, editTokenURI bool, name, description, image, tokenURI string)
}

var _ NFT = (*BaseNFT)(nil)

// BaseNFT non fungible token definition
type BaseNFT struct {
	ID          uint64         `json:"id,omitempty"` // id of the token; not exported to clients
	Owner       sdk.AccAddress `json:"owner"`        // account address that owns the NFT
	Name        string         `json:"name"`         // name of the token
	Description string         `json:"description"`  // unique description of the NFT
	Image       string         `json:"image"`        // image path
	TokenURI    string         `json:"token_uri"`    // optional extra data available fo querying
}

// NewBaseNFT creates a new NFT instance
func NewBaseNFT(ID uint64, owner sdk.AccAddress, tokenURI, description, image, name string,
) BaseNFT {
	return BaseNFT{
		ID:          ID,
		Owner:       owner,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Image:       strings.TrimSpace(image),
		TokenURI:    strings.TrimSpace(tokenURI),
	}
}

// EditMetadata edits metadata of an nft
func (nft *BaseNFT) EditMetadata(editName, editDescription, editImage, editTokenURI bool,
	name, description, image, tokenURI string) {
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
}

func (nft BaseNFT) String() string {
	return fmt.Sprintf(`	ID: 					%d
	Owner:        			%s
  	Name:         			%s
  	Description: 			%s
  	Image:        			%s
	TokenURI:   			%s`,
		nft.ID,
		nft.Owner,
		nft.Name,
		nft.Description,
		nft.Image,
		nft.TokenURI,
	)
}

// ----------------------------------------------------------------------------
// NFT
// TODO: create interface and types for mintable NFT

// NFTs define a list of NFT
type NFTs []NFT

// NewNFTs creates a new set of NFTs
func NewNFTs(nfts ...NFT) NFTs {
	if len(nfts) == 0 {
		return NFTs{}
	}
	return NFTs(nfts)
}

// Add appends two sets of NFTs
func (nfts *NFTs) Add(nftsB NFTs) {
	(*nfts) = append((*nfts), nftsB...)
}

// Delete deletes NFTs from the set
func (nfts *NFTs) Delete(ids ...uint64) error {
	newNFTs, err := removeNFT(*nfts, ids)
	if err != nil {
		return err
	}
	(*nfts) = newNFTs
	return nil
}

// String follows stringer interface
func (nfts NFTs) String() string {
	if len(nfts) == 0 {
		return ""
	}

	out := ""
	for _, nft := range nfts {
		out += fmt.Sprintf("%v\n", nft.String())
	}
	return out[:len(out)-1]
}

// Empty returns true if there are no NFTs and false otherwise.
func (nfts NFTs) Empty() bool {
	return len(nfts) == 0
}

// removeNFT removes NFTs from the set matching the given ids
func removeNFT(nfts NFTs, ids []uint64) (NFTs, error) {
	// TODO: do this efficciently
	return nfts, nil
}

// ----------------------------------------------------------------------------
// Encoding

// NFTJSON is the exported NFT format for clients
type NFTJSON map[uint64]NFT

// MarshalJSON for NFTs
func (nfts NFTs) MarshalJSON() ([]byte, error) {
	nftJSON := make(NFTJSON)

	for _, nft := range nfts {
		id := nft.ID
		// set the pointer of the ID to nil
		ptr := reflect.ValueOf(nft.ID).Elem()
		ptr.Set(reflect.Zero(ptr.Type()))
		nftJSON[id] = nft
	}

	return json.Marshal(nftJSON)
}

// UnmarshalJSON for NFTs
func (nfts *NFTs) UnmarshalJSON(b []byte) error {
	nftJSON := make(NFTJSON)

	if err := json.Unmarshal(b, &nftJSON); err != nil {
		return err
	}

	for id, nft := range nftJSON {
		(*nfts) = append((*nfts), NewNFT(id, nft.Owner, nft.TokenURI, nft.Description, nft.Image, nft.Name))
	}

	return nil
}
