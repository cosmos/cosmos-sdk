package fee_delegation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Defines delegation module constants
const (
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

// FeeAllowance defines a permission for one account to use another account's balance
// to pay fees
type FeeAllowance interface {
	// Accept checks whether this allowance allows the provided fees to be spent,
	// and optionally updates the allowance or deletes it entirely
	Accept(fee sdk.Coins, block abci.Header) (allow bool, updated FeeAllowance, remove bool)
}
