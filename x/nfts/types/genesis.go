package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OwnerNFTs used for import / export user NFT balances via genesis JSON
type OwnerNFTs struct {
	Owner   sdk.AccAddress `json:"owner"`
	Balance Collections    `json:"balance"`
}

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	Balances    []OwnerNFTs `json:"balances"`
	Collections Collections `json:"collections"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(balances []OwnerNFTs, collections Collections) GenesisState {
	return GenesisState{
		Balances:    balances,
		Collections: collections,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState([]OwnerNFTs{}, NewCollections())
}

// ValidateGenesis performs basic validation of nfts genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return nil
}
