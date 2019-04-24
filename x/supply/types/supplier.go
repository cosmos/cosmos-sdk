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
	CirculatingSupply    sdk.Coins `json:"circulating_supply"`     // supply held by accounts that's not vesting
	InitialVestingSupply sdk.Coins `json:"initial_vesting_supply"` // initial locked supply held by vesting accounts
	ModulesSupply        sdk.Coins `json:"modules_supply"`         // supply held by modules acccounts
	TotalSupply          sdk.Coins `json:"total_supply"`           // supply held by modules acccounts
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

	return NewSupplier(sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, sdk.Coins{})
}

// Inflate adds coins to a given supply type and updates the total supply
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

// Deflate safe subtracts coins for a given supply and updates the total supply
func (supplier *Supplier) Deflate(supplyType string, amount sdk.Coins) {

	switch supplyType {

	case TypeCirculating:
		newSupply, ok := supplier.CirculatingSupply.SafeSub(amount)
		if !ok {
			panic(fmt.Sprintf(
				"circulating supply should be greater than given amount: %s < %s",
				supplier.CirculatingSupply.String(), amount.String(),
			))
		}
		supplier.CirculatingSupply = newSupply

	case TypeModules:
		newSupply, ok := supplier.ModulesSupply.SafeSub(amount)
		if !ok {
			panic(fmt.Sprintf(
				"modules supply should be greater than given amount: %s < %s",
				supplier.ModulesSupply.String(), amount.String(),
			))
		}
		supplier.ModulesSupply = newSupply

	case TypeTotal:
		newSupply, ok := supplier.TotalSupply.SafeSub(amount)
		if !ok {
			panic(fmt.Sprintf(
				"total supply should be greater than given amount: %s < %s",
				supplier.TotalSupply.String(), amount.String(),
			))
		}
		supplier.TotalSupply = newSupply

	default:
		panic(fmt.Errorf("invalid type %s", supplyType))
	}
}

// CirculatingAmountOf returns the circulating supply of a coin denomination
func (supplier Supplier) CirculatingAmountOf(denom string) sdk.Int {

	return supplier.CirculatingSupply.AmountOf(denom)
}

// VestingAmountOf returns the vesting supply of a coin denomination
func (supplier Supplier) InitalVestingAmountOf(denom string) sdk.Int {

	return supplier.InitialVestingSupply.AmountOf(denom)
}

// ModulesAmountOf returns the total token holders' supply of a coin denomination
func (supplier Supplier) ModulesAmountOf(denom string) sdk.Int {

	return supplier.ModulesSupply.AmountOf(denom)
}

// TotalAmountOf returns the sum of circulating, vesting and modules supply for a specific coin denomination
func (supplier Supplier) TotalAmountOf(denom string) sdk.Int {

	return supplier.TotalSupply.AmountOf(denom)
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

	return nil
}

// String returns a human readable string representation of a supplier.
func (supplier Supplier) String() string {
	return fmt.Sprintf(`Supplier:
  Circulating Supply:  %s
  Initial Vesting Supply: %s
	Modules Supply:  %s
	Total Supply:  %s`,
		supplier.CirculatingSupply.String(),
		supplier.InitialVestingSupply.String(),
		supplier.ModulesSupply.String(),
		supplier.TotalSupply.String())
}
