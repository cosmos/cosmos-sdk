package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// define constants for inflation
const (
	TypeCirculating = "circulating"
	TypeVesting     = "vesting"
	TypeModules     = "modules"
	TypeTotal       = "total"
)

// Supplier represents a struct that passively keeps track of the total supply amounts in the network
type Supplier struct {
	CirculatingSupply    sdk.Coins `json:"circulating_supply"`     // supply held by accounts that's not vesting; circulating = total - vesting
	InitialVestingSupply sdk.Coins `json:"initial_vesting_supply"` // initial locked supply held by vesting accounts
	ModulesSupply        sdk.Coins `json:"modules_supply"`         // supply held by modules acccounts
	TotalSupply          sdk.Coins `json:"total_supply"`           // total supply of tokens on the chain
}

// NewSupplier creates a new Supplier instance
func NewSupplier(circulating, vesting, modules, total sdk.Coins) Supplier {

	return Supplier{
		CirculatingSupply:    circulating,
		InitialVestingSupply: vesting,
		ModulesSupply:        modules,
		TotalSupply:          total,
	}
}

// DefaultSupplier creates an empty Supplier
func DefaultSupplier() Supplier {
	coins := sdk.NewCoins()
	return NewSupplier(coins, coins, coins, coins)
}

// Inflate adds coins to a given supply type
func (supplier *Supplier) Inflate(supplyType string, amount sdk.Coins) {
	switch supplyType {

	case TypeCirculating:
		supplier.CirculatingSupply = supplier.CirculatingSupply.Add(amount)

	case TypeVesting:
		supplier.InitialVestingSupply = supplier.InitialVestingSupply.Add(amount)

	case TypeModules:
		supplier.ModulesSupply = supplier.ModulesSupply.Add(amount)

	case TypeTotal:
		supplier.TotalSupply = supplier.TotalSupply.Add(amount)

	default:
		panic(fmt.Errorf("invalid type %s", supplyType))
	}
}

// Deflate subtracts coins for a given supply
func (supplier *Supplier) Deflate(supplyType string, amount sdk.Coins) {

	switch supplyType {

	case TypeCirculating:
		supplier.CirculatingSupply = supplier.CirculatingSupply.Sub(amount)

	case TypeModules:
		supplier.ModulesSupply = supplier.ModulesSupply.Sub(amount)

	case TypeTotal:
		supplier.TotalSupply = supplier.TotalSupply.Sub(amount)

	default:
		panic(fmt.Errorf("invalid type %s", supplyType))
	}
}

// ValidateBasic validates the Supply coins and returns error if invalid
func (supplier Supplier) ValidateBasic() sdk.Error {

	if !supplier.CirculatingSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid circulating supply: %s", supplier.CirculatingSupply.String()),
		)
	}
	if !supplier.InitialVestingSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid initial vesting supply: %s", supplier.InitialVestingSupply.String()),
		)
	}
	if !supplier.ModulesSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid token holders supply: %s", supplier.ModulesSupply.String()),
		)
	}
	if !supplier.TotalSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid total supply: %s", supplier.ModulesSupply.String()),
		)
	}

	return nil
}

// String returns a human readable string representation of a supplier.
func (supplier Supplier) String() string {
	return fmt.Sprintf(`Supplier:
  Circulating Supply:     %s
  Initial Vesting Supply: %s
	Modules Supply:         %s
	Total Supply:           %s`,
		supplier.CirculatingSupply.String(),
		supplier.InitialVestingSupply.String(),
		supplier.ModulesSupply.String(),
		supplier.TotalSupply.String())
}
