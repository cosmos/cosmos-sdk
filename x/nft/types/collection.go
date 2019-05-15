package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Collection of non fungible tokens
type Collection struct {
	Denom string `json:"denom,omitempty"` // name of the collection; not exported to clients
	NFTs  NFTs   `json:"nfts"`            // NFTs that belong to a collection
}

// NewCollection creates a new NFT Collection
func NewCollection(denom string, nfts NFTs) Collection {
	return Collection{
		Denom: strings.TrimSpace(denom),
		NFTs:  NewNFTs([]NFT(nfts)...),
	}
}

// EmptyCollection returns an empty collection
func EmptyCollection() Collection {
	return NewCollection("", NewNFTs())
}

// GetNFT gets a NFT from the collection
func (collection Collection) GetNFT(id uint64) (nft NFT, err sdk.Error) {

	for _, nft := range collection.NFTs {
		if nft.GetID() == id {
			return nft, nil
		}
	}

	return nil, ErrUnknownNFT(DefaultCodespace,
		fmt.Sprintf("NFT #%d doesn't exist in collection %s", nft.GetID(), collection.Denom),
	)
}

// AddNFT adds an NFT to the collection
func (collection *Collection) AddNFT(nft NFT) {
	collection.NFTs = append(collection.NFTs, nft)
}

// DeleteNFT deletes an NFT from a collection
func (collection *Collection) DeleteNFT(nft NFT) sdk.Error {
	nfts, ok := collection.NFTs.Remove(nft.GetID())
	if !ok {
		ErrUnknownNFT(DefaultCodespace,
			fmt.Sprintf("NFT #%d doesn't exist on collection %s", nft.GetID(), collection.Denom),
		)
	}
	(*collection).NFTs = nfts
	return nil
}

// Supply gets the total supply of NFTs of a collection
func (collection Collection) Supply() uint {
	return uint(len(collection.NFTs))
}

// String follows stringer interface
func (collection Collection) String() string {
	return fmt.Sprintf(`
	Denom: 				%s
	NFTs:        	%s`,
		collection.Denom,
		collection.NFTs.String(),
	)
}

// ----------------------------------------------------------------------------
// Collections

// Collections define an array of Collection
type Collections []Collection

// NewCollections creates a new set of NFTs
func NewCollections(collections ...Collection) Collections {
	if len(collections) == 0 {
		return Collections{}
	}
	return Collections(collections)
}

// Add appends two sets of Collections
func (collections *Collections) Add(collectionsB Collections) {
	(*collections) = append((*collections), collectionsB...)
}

// Find returns the searched collection from the set
func (collections Collections) Find(denom string) (Collection, bool) {
	index := collections.find(denom)
	if index == -1 {
		return Collection{}, false
	}
	return collections[index], true
}

// Remove removes a collection from the set of collections
func (collections Collections) Remove(denom string) (Collections, bool) {
	index := collections.find(denom)
	if index == -1 {
		return collections, false
	}

	return append(collections[:index], collections[:index+1]...), true
}

// String follows stringer interface
func (collections Collections) String() string {
	if len(collections) == 0 {
		return ""
	}

	out := ""
	for _, collection := range collections {
		out += fmt.Sprintf("%v\n", collection.String())
	}
	return out[:len(out)-1]
}

// Empty returns true if there are no collections and false otherwise.
func (collections Collections) Empty() bool {
	return len(collections) == 0
}

func (collections Collections) find(denom string) uint {
	if len(collections) == 0 {
		return -1
	}

	midIdx := uint(len(collections)) / 2
	midCollection := collections[midIdx]

	if strings.Compare(denom, midCollection.Denom) == -1 {
		return collections[:midIdx].find(denom)
	} else if midCollection.Denom == denom {
		return midIdx
	} else {
		return collections[midIdx+1:].find(denom)
	}
}

// ----------------------------------------------------------------------------
// Encoding

// CollectionJSON is the exported Collection format for clients
type CollectionJSON map[string]Collection

// MarshalJSON for Collections
func (collections Collections) MarshalJSON() ([]byte, error) {
	collectionJSON := make(CollectionJSON)

	for _, collection := range collections {
		denom := collection.Denom
		// set the pointer of the ID to nil
		ptr := reflect.ValueOf(collection.Denom).Elem()
		ptr.Set(reflect.Zero(ptr.Type()))
		collectionJSON[denom] = collection
	}

	return json.Marshal(collectionJSON)
}

// UnmarshalJSON for Collections
func (collections *Collections) UnmarshalJSON(b []byte) error {
	collectionJSON := make(CollectionJSON)

	if err := json.Unmarshal(b, &collectionJSON); err != nil {
		return err
	}

	for denom, collection := range collectionJSON {
		*collections = append(*collections, NewCollection(denom, collection.NFTs))
	}

	return nil
}

//-----------------------------------------------------------------------------
// Sort interface

//nolint
func (collections Collections) Len() int { return len(collections) }
func (collections Collections) Less(i, j int) bool {
	return strings.Compare(collections[i].Denom, collections[j].Denom) == -1
}
func (collections Collections) Swap(i, j int) {
	collections[i], collections[j] = collections[j], collections[i]
}

var _ sort.Interface = Collections{}

// Sort is a helper function to sort the set of coins inplace
func (collections Collections) Sort() Collections {
	sort.Sort(collections)
	return collections
}
