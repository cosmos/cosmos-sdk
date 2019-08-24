package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IDCollection defines a set of nft ids that belong to a specific
// collection
type IDCollection struct {
	Denom string   `json:"denom" yaml:"denom"`
	IDs   []string `json:"ids" yaml:"ids"`
}

// NewIDCollection creates a new IDCollection instance
func NewIDCollection(denom string, ids []string) IDCollection {
	return IDCollection{
		Denom: strings.TrimSpace(denom),
		IDs:   ids,
	}
}

// Exists determines whether an ID is in the IDCollection
func (idCollection IDCollection) Exists(id string) (exists bool) {
	// TODO: improve performance
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
	index := stringArray(idCollection.IDs).find(id)
	if index == -1 {
		return idCollection, ErrUnknownNFT(DefaultCodespace,
			fmt.Sprintf("ID #%s doesn't exist on ID Collection %s", id, idCollection.Denom),
		)
	}

	idCollection.IDs = append(idCollection.IDs[:index], idCollection.IDs[index+1:]...)

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

	switch {
	case strings.Compare(el, midIDCollection.Denom) == -1:
		return idCollections[:midIdx].find(el)
	case midIDCollection.Denom == el:
		return midIdx
	default:
		return idCollections[midIdx+1:].find(el)
	}
}

// Owner of non fungible tokens
type Owner struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	IDCollections IDCollections  `json:"idCollections" yaml:"idCollections"`
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
		return IDCollection{}, false
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

	owner.IDCollections = append(append(owner.IDCollections[:index], idCollection), owner.IDCollections[index+1:]...)

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

// stringArray is an array of strings whose sole purpose is to help with find
type stringArray []string

func (sa stringArray) find(el string) (idx int) {
	if len(sa) == 0 {
		return -1
	}

	midIdx := len(sa) / 2
	stringArrayEl := sa[midIdx]

	switch {
	case strings.Compare(el, stringArrayEl) == -1:
		return sa[:midIdx].find(el)
	case stringArrayEl == el:
		return midIdx
	default:
		return sa[midIdx+1:].find(el)
	}
}
