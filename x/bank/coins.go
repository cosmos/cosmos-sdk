package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DenomMetadata struct {
	Name        string
	Symbol      string
	Decimals    int8
	TotalSupply sdk.Int
}
