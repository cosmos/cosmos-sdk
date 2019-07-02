package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IDCollection of non fungible tokens
type IDCollection struct {
	Denom string   `json:"denom"`
	IDs   []string `json:"IDs"`
}

// StringArray is an array of strings whose sole purpose is to help with find
type StringArray []string

func (stringArray StringArray) find(el string) (idx int) {
	if len(stringArray) == 0 {
		return -1
	}
	midIdx := len(stringArray) / 2
	stringArrayEl := stringArray[midIdx]

	if strings.Compare(el, stringArrayEl) == -1 {
		return stringArray[:midIdx].find(el)
	} else if stringArrayEl == el {
		return midIdx
	} else {
		return stringArray[midIdx+1:].find(el)
	}
}

// NewIDCollection creates a new NFT NFTIDs
func NewIDCollection(denom string, ids []string) IDCollection {
	return IDCollection{
		Denom: strings.TrimSpace(denom),
		IDs:   ids,
	}
}

// EmptyIDCollection returns an empty balance
func EmptyIDCollection() IDCollection {
	return NewIDCollection("", []string{})
}

// Exists determines whether an ID is in the IDCollection
func (idCollection IDCollection) Exists(id string) (exists bool) {
	for _, _id := range idCollection.IDs {
		if _id == id {
			return true
		}
	}
	return false
}

// AddID adds an ID to the idCollection
func (idCollection IDCollection) AddID(id string) IDCollection {
	idCollection.IDs = append(idCollection.IDs, id)
	return idCollection
}

// DeleteID deletes an ID from an ID Collection
func (idCollection IDCollection) DeleteID(id string) (IDCollection, sdk.Error) {
	index := StringArray(idCollection.IDs).find(id)
	if index == -1 {
		return idCollection, ErrUnknownNFT(DefaultCodespace,
			fmt.Sprintf("ID #%s doesn't exist on ID Collection %s", id, idCollection.Denom),
		)
	}
	idCollection.IDs[len(idCollection.IDs)-1], idCollection.IDs[index] = idCollection.IDs[index], idCollection.IDs[len(idCollection.IDs)-1]
	idCollection.IDs = idCollection.IDs[:len(idCollection.IDs)-1]
	return idCollection, nil
}

// Supply gets the total supply of NFTIDs of a balance
func (idCollection IDCollection) Supply() int {
	return len(idCollection.IDs)
}

// String follows stringer interface
func (idCollection IDCollection) String() string {
	return fmt.Sprintf(`Denom: 			%s
IDs:        	%s`,
		idCollection.Denom,
		strings.Join(idCollection.IDs, ","),
	)
}

// ----------------------------------------------------------------------------
// Owners

// IDCollections is an array of ID Collections whose sole purpose is for find
type IDCollections []IDCollection

// String follows stringer interface
func (idCollections IDCollections) String() string {
	if len(idCollections) == 0 {
		return ""
	}

	out := ""
	for _, idCollection := range idCollections {
		out += fmt.Sprintf("%v\n", idCollection.String())
	}
	return out[:len(out)-1]
}

func (idCollections IDCollections) find(el string) int {
	if len(idCollections) == 0 {
		return -1
	}

	midIdx := len(idCollections) / 2
	midIDCollection := idCollections[midIdx]

	if strings.Compare(el, midIDCollection.Denom) == -1 {
		return idCollections[:midIdx].find(el)
	} else if midIDCollection.Denom == el {
		return midIdx
	} else {
		return idCollections[midIdx+1:].find(el)
	}
}

// Owner of non fungible tokens
type Owner struct {
	Address       sdk.AccAddress `json:"address"`
	IDCollections IDCollections  `json:"IDCollections"`
}

// NewOwner creates a new Owner
func NewOwner(owner sdk.AccAddress, idCollections ...IDCollection) Owner {
	return Owner{
		Address:       owner,
		IDCollections: idCollections,
	}
}

// Supply gets the total supply of an Owner
func (owner Owner) Supply() int {
	total := 0
	for _, idCollection := range owner.IDCollections {
		total += idCollection.Supply()
	}
	return total
}

// GetIDCollection gets the IDCollection from the owner
func (owner Owner) GetIDCollection(denom string) (IDCollection, bool) {
	index := owner.IDCollections.find(denom)
	if index == -1 {
		return EmptyIDCollection(), false
	}
	return owner.IDCollections[index], true
}

// UpdateIDCollection updates the ID Collection of an owner
func (owner Owner) UpdateIDCollection(idCollection IDCollection) (Owner, sdk.Error) {
	denom := idCollection.Denom
	index := owner.IDCollections.find(denom)
	if index == -1 {
		return owner, ErrUnknownCollection(DefaultCodespace,
			fmt.Sprintf("ID Collection %s doesn't exist for owner %s", denom, owner.Address),
		)
	}
	owner.IDCollections = append(append(owner.IDCollections[:index], idCollection), owner.IDCollections[:index+1]...)
	return owner, nil
}

// DeleteID deletes an ID from an owners ID Collection
func (owner Owner) DeleteID(denom string, id string) (Owner, sdk.Error) {
	idCollection, found := owner.GetIDCollection(denom)
	if !found {
		return owner, ErrUnknownNFT(DefaultCodespace,
			fmt.Sprintf("ID #%s doesn't exist in ID Collection %s", id, denom),
		)
	}
	idCollection, err := idCollection.DeleteID(id)
	if err != nil {
		return owner, err
	}
	owner, err = owner.UpdateIDCollection(idCollection)
	if err != nil {
		return owner, err
	}
	return owner, nil
}

// String follows stringer interface
func (owner Owner) String() string {
	return fmt.Sprintf(`
	Address: 				%s
	IDCollections:        	%s`,
		owner.Address,
		owner.IDCollections.String(),
	)
}

// // String follows stringer interface
// func (balances Owners) String() string {
// 	if len(balances) == 0 {
// 		return ""
// 	}

// 	out := ""
// 	for _, balance := range balances {
// 		out += fmt.Sprintf("%v\n", balance.String())
// 	}
// 	return out[:len(out)-1]
// }

// // Empty returns true if there are no balances and false otherwise.
// func (balances Owners) Empty() bool {
// 	return len(balances) == 0
// }

// ----------------------------------------------------------------------------
// Encoding

// // UnmarshalJSON for Owners
// func (owner *Owner) UnmarshalJSON(b []byte) error {
// 	idCollectionJSON := make(IDCollectionJSON)

// 	if err := json.Unmarshal(b, &idCollectionJSON); err != nil {
// 		return err
// 	}

// 	var idCollections []IDCollection

// 	for denom, idCollection := range idCollectionJSON {
// 		*owner.IDCollections = append(*owner.IDCollections, NewIDCollection(denom, idCollection))
// 	}

// 	return nil
// }

// // UnmarshalJSON for Collections
// func (collections *Collections) UnmarshalJSON(b []byte) error {
// 	collectionJSON := make(CollectionJSON)

// 	if err := json.Unmarshal(b, &collectionJSON); err != nil {
// 		return err
// 	}

// 	for denom, collection := range collectionJSON {
// 		*collections = append(*collections, NewCollection(denom, collection.NFTs))
// 	}

// 	return nil
// }

// //-----------------------------------------------------------------------------
// // Sort interface

// //nolint
// func (balances Owners) Len() int { return len(balances) }
// func (balances Owners) Less(i, j int) bool {
// 	return strings.Compare(balances[i].Denom, balances[j].Denom) == -1
// }
// func (balances Owners) Swap(i, j int) {
// 	balances[i], balances[j] = balances[j], balances[i]
// }

// var _ sort.Interface = Owners{}

// // Sort is a helper function to sort the set of coins inplace
// func (balances Owners) Sort() Owners {
// 	sort.Sort(balances)
// 	return balances
// }
