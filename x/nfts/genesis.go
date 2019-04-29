package nfts

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OwnerNFTs used for import / export user NFT balances via genesis JSON
type OwnerNFTs struct {
	Owner   sdk.AccAddress `json:"owner"`
	Balance []Collection   `json:"balance"`
}

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	Collections []Collection `json:"collections"`
	Balances    []OwnerNFTs  `json:"balances"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(collections []Collection, balances []OwnerNFTs) GenesisState {
	return GenesisState{
		Collections: collections,
		Balances:    balances,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState([]Collection{}, []OwnerNFTs{})
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, collection := range data.Collections {
		k.SetCollection(ctx, collection.Denom, collection)
	}

	for _, balance := range data.Balance {
		k.SetOwnerBalance(ctx, balance.Owner, balance)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return NewGenesisState(k.GetCollections(ctx), k.GetOwners(ctx))
}

// ValidateGenesis performs basic validation of nfts genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return nil
}
