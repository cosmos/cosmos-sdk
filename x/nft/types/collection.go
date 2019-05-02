package types

import (
	"encoding/json"
	"fmt"
	"reflect"
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

// Empty returns true if there are no collections and false otherwise.
func (collections Collections) Empty() bool {
	return len(collections) == 0
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
		(*collections) = append((*collections), NewCollection(denom, collection.NFTs))
	}

	return nil
}
