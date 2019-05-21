package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NFT non fungible token interface
type NFT interface {
	GetID() uint64
	GetOwner() sdk.AccAddress
	SetOwner(address sdk.AccAddress) BaseNFT
	GetName() string
	GetDescription() string
	GetImage() string
	GetTokenURI() string

	EditMetadata(name, description, image, tokenURI string) BaseNFT
	String() string
}

var _ NFT = (*BaseNFT)(nil)

// BaseNFT non fungible token definition
type BaseNFT struct {
	ID          uint64         `json:"id,omitempty"` // id of the token; not exported to clients
	Owner       sdk.AccAddress `json:"owner"`        // account address that owns the NFT
	Name        string         `json:"name"`         // name of the token
	Description string         `json:"description"`  // unique description of the NFT
	Image       string         `json:"image"`        // image path
	TokenURI    string         `json:"token_uri"`    // optional extra properties available fo querying
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

// GetID returns the ID of the token
func (bnft BaseNFT) GetID() uint64 { return bnft.ID }

// GetOwner returns the account address that owns the NFT
func (bnft BaseNFT) GetOwner() sdk.AccAddress { return bnft.Owner }

// SetOwner updates the owner address of the NFT
func (bnft BaseNFT) SetOwner(address sdk.AccAddress) BaseNFT {
	return BaseNFT{
		ID:          bnft.GetID(),
		Owner:       address,
		Name:        bnft.GetName(),
		Description: bnft.GetDescription(),
		Image:       bnft.GetImage(),
		TokenURI:    bnft.GetTokenURI(),
	}
}

// GetName returns the name of the token
func (bnft BaseNFT) GetName() string { return bnft.Name }

// GetDescription returns the unique description of the NFT
func (bnft BaseNFT) GetDescription() string { return bnft.Description }

// GetImage returns the image path of the NFT
func (bnft BaseNFT) GetImage() string { return bnft.Image }

// GetTokenURI returns the path to optional extra properties
func (bnft BaseNFT) GetTokenURI() string { return bnft.TokenURI }

// EditMetadata edits metadata of an nft
func (bnft BaseNFT) EditMetadata(name, description, image, tokenURI string) BaseNFT {
	return BaseNFT{
		ID:          bnft.GetID(),
		Owner:       bnft.GetOwner(),
		Name:        name,
		Description: description,
		Image:       image,
		TokenURI:    tokenURI,
	}
}

func (bnft BaseNFT) String() string {
	return fmt.Sprintf(`ID:				%d
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`,
		bnft.ID,
		bnft.Owner,
		bnft.Name,
		bnft.Description,
		bnft.Image,
		bnft.TokenURI,
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
	*nfts = append(*nfts, nftsB...)
}

// Find returns the searched collection from the set
func (nfts NFTs) Find(id uint64) (nft NFT, found bool) {
	index := nfts.find(id)
	if index == -1 {
		return nft, false
	}
	return nfts[index], true
}

// Update removes and replaces an NFT from the set
func (nfts NFTs) Update(id uint64, nft NFT) (NFTs, bool) {
	index := nfts.find(id)
	if index == -1 {
		return nfts, false
	}

	return append(append(nfts[:index], nft), nfts[:index+1]...), true
}

// Remove removes a collection from the set of collections
func (nfts NFTs) Remove(id uint64) (NFTs, bool) {
	index := nfts.find(id)
	if index == -1 {
		return nfts, false
	}

	return append(nfts[:index], nfts[:index+1]...), true
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

func (nfts NFTs) find(id uint64) int {
	if len(nfts) == 0 {
		return -1
	}

	midIdx := len(nfts) / 2
	nft := nfts[midIdx]

	if id < nft.GetID() {
		return nfts[:midIdx].find(id)
	} else if id == nft.GetID() {
		return midIdx
	} else {
		return nfts[midIdx+1:].find(id)
	}
}

// ----------------------------------------------------------------------------
// Encoding

// NFTJSON is the exported NFT format for clients
type NFTJSON map[uint64]NFT

// MarshalJSON for NFTs
func (nfts NFTs) MarshalJSON() ([]byte, error) {
	nftJSON := make(NFTJSON)

	for _, nft := range nfts {
		id := nft.GetID()
		// set the pointer of the ID to nil
		ptr := reflect.ValueOf(id).Elem()
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
		*nfts = append(*nfts, NewBaseNFT(id, nft.GetOwner(), nft.GetTokenURI(), nft.GetDescription(), nft.GetImage(), nft.GetName()))
	}

	return nil
}

//-----------------------------------------------------------------------------
// Sort interface

//nolint
func (nfts NFTs) Len() int           { return len(nfts) }
func (nfts NFTs) Less(i, j int) bool { return nfts[i].GetID() < nfts[j].GetID() }
func (nfts NFTs) Swap(i, j int)      { nfts[i], nfts[j] = nfts[j], nfts[i] }

var _ sort.Interface = NFTs{}

// Sort is a helper function to sort the set of coins inplace
func (nfts NFTs) Sort() NFTs {
	sort.Sort(nfts)
	return nfts
}
