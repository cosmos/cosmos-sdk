package nfts

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	Collections map[Denom]Collection `json:"collections"`
	Owners      map[string]Owner     `json:"owners"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(collections map[Denom]Collection, owners map[string]Owner) GenesisState {
	return GenesisState{
		Collections: collections,
		Owners:      owners,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(map[Denom]Collection{}, map[string]Owner{})
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	for denom, collection := range data.Collections {
		keeper.SetCollection(ctx, denom, collection)
		for id, nft := range collection {
			keeper.AddToOwner(ctx, denom, id, nft)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(keeper.GetCollections(ctx), keeper.GetOwners(ctx))
}

// ValidateGenesis performs basic validation of nfts genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	for denom, collection := range data.Collections {
		var err sdk.Error
		if strings.TrimSpace(string(denom)) == "" {
			err = ErrInvalidNFT(DefaultCodespace)
		} else {
			err = collection.ValidateBasic()
		}
		if err != nil {
			return err
		}
	}
	return nil
}
