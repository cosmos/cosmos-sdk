package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Balance used for import / export user NFT balances via genesis JSON
type Balance struct {
	Collection
	Owner sdk.AccAddress `json:"owner"`
}

// NewBalance creates a new NFT balance instance
func NewBalance(collection Collection, owner sdk.AccAddress) Balance {
	return Balance{collection, owner}
}

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Balances []Balance `json:"balances"`
	Collections
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(balances []Balance, collections Collections) GenesisState {
	return GenesisState{
		Balances:    balances,
		Collections: collections,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState([]Balance{}, NewCollections())
}

// ValidateGenesis performs basic validation of nfts genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return nil
}
