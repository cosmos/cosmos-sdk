package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Type that stores all the important metadata related to a denom
type DenomMetadata struct {
	Name        string
	Symbol      string
	Decimals    int8
	TotalSupply sdk.Int
}
