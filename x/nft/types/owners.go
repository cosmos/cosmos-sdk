package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Owned struct {
	Address sdk.AccAddress
	Owned   []NFTIDs
}

// NFTIDs of non fungible tokens
type NFTIDs struct {
	Denom  string   `json:"denom,string,omitempty"`
	NFTIDs []uint64 `json:"NFTIDs"`
}

// // NewNFTIDs creates a new NFT NFTIDs
// func NewNFTIDs(denom string, nfts []uint64]) NFTIDs {
// 	return NFTIDs{
// 		Denom: strings.TrimSpace(denom),
// 		NFTIDs:  nfts
// 	}
// }

// // EmptyNFTIDs returns an empty balance
// func EmptyNFTIDs() NFTIDs {
// 	return NewNFTIDs("", []uint64)
// }

// // GetNFT gets a NFT from the balance
// func (balance NFTIDs) GetNFT(id uint64) (exists bool) {

// 	for _, nft := range balance.NFTIDs {
// 		if nft == id {
// 			return true
// 		}
// 	}
// 	return false
// }

// // AddNFT adds an NFT to the balance
// func (balance *NFTIDs) AddNFT(nft uint64) {
// 	balance.NFTIDs = append(balance.NFTIDs, nft)
// }

// // DeleteNFT deletes an NFT from a balance
// func (balance *NFTIDs) DeleteNFT(nft uint64) sdk.Error {
// 	// TODO: how to remove element from array
// 	// nfts, ok := balance.NFTIDs.Remove(nft)
// 	if !ok {
// 		return ErrUnknownNFT(DefaultCodespace,
// 			fmt.Sprintf("NFT #%d doesn't exist on balance %s", nft.GetID(), balance.Denom),
// 		)
// 	}
// 	(*balance).NFTIDs = nfts
// 	return nil
// }

// // Supply gets the total supply of NFTIDs of a balance
// func (balance NFTIDs) Supply() uint {
// 	return uint(len(balance.NFTIDs))
// }

// // String follows stringer interface
// func (balance NFTIDs) String() string {
// 	return fmt.Sprintf(`
// 	Denom: 				%s
// 	NFTIDs:        	%s`,
// 		balance.Denom,
// 		balance.NFTIDs.String(), // TODO: check array to string
// 	)
// }

// // ----------------------------------------------------------------------------
// // Owners

// // Owners an owner's array of various NFT IDs by denom
// type Owners struct {
// 	Owner sdk.AccAddress `json:"owner,string"`
// 	Owners []Owner  `json:"Owners"`
// }

// // NewOwners creates a new set of NFTIDs
// func NewOwners(owner sdk.AccAddress, balances ...Owner) Owners {
// 	return Owner{
// 		Owner: owner,
// 		Owners:  balances,
// 	}
// }

// // Add appends two sets of Owners
// func (balances *Owners) Add(balancesB Owners) {
// 	*balances = append(*balances, balancesB...)
// }

// // Find returns the searched balance from the set
// func (balances Owners) Find(owner sdk.AccAddress, denom string) (Owner, bool) {
// 	index := balances.find(owner)
// 	if index == -1 {
// 		return Owner{}, false
// 	}
// 	return balances[index], true
// }

// // Remove removes a balance from the set of balances
// func (balances Owners) Remove(denom string) (Owners, bool) {
// 	index := balances.find(denom)
// 	if index == -1 {
// 		return balances, false
// 	}

// 	return append(balances[:index], balances[:index+1]...), true
// }

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

// func (balances Owners) find(owner string) int {
// 	if len(balances) == 0 {
// 		return -1
// 	}

// 	midIdx := len(balances) / 2
// 	midOwner := balances[midIdx]

// 	if strings.Compare(denom, midOwner.Denom) == -1 {
// 		return balances[:midIdx].find(denom)
// 	} else if midOwner.Denom == denom {
// 		return midIdx
// 	} else {
// 		return balances[midIdx+1:].find(denom)
// 	}
// }

// // ----------------------------------------------------------------------------
// // Encoding

// // OwnerJSON is the exported Owner format for clients
// type OwnerJSON map[string]Owner

// // MarshalJSON for Owners
// func (balances Owners) MarshalJSON() ([]byte, error) {
// 	balanceJSON := make(OwnerJSON)

// 	for _, balance := range balances {
// 		denom := balance.Denom
// 		// set the pointer of the ID to nil
// 		ptr := reflect.ValueOf(balance.Denom).Elem()
// 		ptr.Set(reflect.Zero(ptr.Type()))
// 		balanceJSON[denom] = balance
// 	}

// 	return json.Marshal(balanceJSON)
// }

// // UnmarshalJSON for Owners
// func (balances *Owners) UnmarshalJSON(b []byte) error {
// 	balanceJSON := make(OwnerJSON)

// 	if err := json.Unmarshal(b, &balanceJSON); err != nil {
// 		return err
// 	}

// 	for denom, balance := range balanceJSON {
// 		*balances = append(*balances, NewOwner(denom, balance.NFTIDs))
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
