package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "nft"
)

func NewCollectionInfo(id, name, schema string, creator sdk.AccAddress) CollectionInfo {
	return CollectionInfo{
		Id:      id,
		Name:    name,
		Schema:  schema,
		Creator: creator.String(),
	}
}
