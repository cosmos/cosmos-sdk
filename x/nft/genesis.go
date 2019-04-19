package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	Collections map[types.Denom]types.Collection `json:"collections"`
	Owners      map[string]types.Owner           `json:"owners"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(collections map[types.Denom]types.Collection, owners map[string]types.Owner) GenesisState {
	return GenesisState{
		Collections: collections,
		Owners:      owners,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(map[types.Denom]types.Collection{}, map[string]types.Owner{})
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data GenesisState) {
	for denom, collection := range data.Collections {
		k.SetCollection(ctx, denom, collection)
		for id, nft := range collection {
			k.AddToOwner(ctx, denom, id, nft)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) GenesisState {
	return NewGenesisState(k.GetCollections(ctx), k.GetOwners(ctx))
}

// ValidateGenesis performs basic validation of nfts genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return nil
}
