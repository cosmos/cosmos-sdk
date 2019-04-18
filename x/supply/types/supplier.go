package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// define consts for inflation
const (
	TypeCirculating = "circulating"
	TypeVesting     = "vesting"
	TypeModules     = "modules"
)

// Supplier represents a struct that passively keeps track of the total supply amounts in the network
type Supplier struct {
	CirculatingSupply sdk.Coins `json:"circulating_supply"` // supply held by accounts that's not vesting
	VestingSupply     sdk.Coins `json:"vesting_supply"`     // locked supply held by vesting accounts
	ModulesSupply     sdk.Coins `json:"modules_supply"`     // supply held by modules acccounts
	TotalSupply       sdk.Coins `json:"total_supply"`       // total supply of the network
}

// NewSupplier creates a new Supplier instance
func NewSupplier(circulating, vesting, modules, total sdk.Coins) Supplier {
	return Supplier{
		CirculatingSupply: circulating,
		VestingSupply:     vesting,
		ModulesSupply:     modules,
		TotalSupply:       total,
	}
}

// DefaultSupplier creates an empty Supplier
func DefaultSupplier() Supplier {
	return NewSupplier(sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, sdk.Coins{})
}

// Inflate adds coins to a given supply type and updates the total supply
func (supplier *Supplier) Inflate(supplyType string, amount sdk.Coins) error {
	switch supplyType {
	case TypeCirculating:
		supplier.CirculatingSupply = supplier.CirculatingSupply.Add(amount)
	case TypeVesting:
		supplier.VestingSupply = supplier.VestingSupply.Add(amount)
	case TypeModules:
		supplier.ModulesSupply = supplier.ModulesSupply.Add(amount)
	default:
		return fmt.Errorf("invalid type %s", supplyType)
	}
	supplier.TotalSupply = supplier.TotalSupply.Add(amount)
	return nil
}

// Deflate safe substracts coins for a given supply and updates the total supply
func (supplier *Supplier) Deflate(supplyType string, amount sdk.Coins) error {
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
	case TypeVesting:
		newSupply, ok := supplier.VestingSupply.SafeSub(amount)
		if !ok {
			panic(fmt.Sprintf(
				"vesting supply should be greater than given amount: %s < %s",
				supplier.VestingSupply.String(), amount.String(),
			))
		}
		supplier.VestingSupply = newSupply
	case TypeModules:
		newSupply, ok := supplier.ModulesSupply.SafeSub(amount)
		if !ok {
			panic(fmt.Sprintf(
				"modules supply should be greater than given amount: %s < %s",
				supplier.ModulesSupply.String(), amount.String(),
			))
		}
		supplier.ModulesSupply = newSupply
	default:
		return fmt.Errorf("invalid type %s", supplyType)
	}

	newSupply, ok := supplier.TotalSupply.SafeSub(amount)
	if !ok {
		panic(fmt.Sprintf(
			"total supply should be greater than given amount: %s < %s",
			supplier.TotalSupply.String(), amount.String(),
		))
	}
	supplier.TotalSupply = newSupply
	return nil
}

// CirculatingAmountOf returns the circulating supply of a coin denomination
func (supplier Supplier) CirculatingAmountOf(denom string) sdk.Int {
	return supplier.CirculatingSupply.AmountOf(denom)
}

// VestingAmountOf returns the vesting supply of a coin denomination
func (supplier Supplier) VestingAmountOf(denom string) sdk.Int {
	return supplier.VestingSupply.AmountOf(denom)
}

// ModulesAmountOf returns the total token holders' supply of a coin denomination
func (supplier Supplier) ModulesAmountOf(denom string) sdk.Int {
	return supplier.ModulesSupply.AmountOf(denom)
}

// TotalAmountOf returns the total supply of a coin denomination
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
	if !supplier.VestingSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid vesting supply: %s", supplier.VestingSupply.String()),
		)
	}
	if !supplier.ModulesSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid token holders supply: %s", supplier.ModulesSupply.String()),
		)
	}
	if !supplier.TotalSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid total supply: %s", supplier.TotalSupply.String()),
		)
	}

	calculatedTotalSupply :=
		supplier.CirculatingSupply.Add(supplier.VestingSupply).Add(supplier.ModulesSupply)

	if !supplier.TotalSupply.IsEqual(calculatedTotalSupply) {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("total supply ≠ calculated total supply: %s ≠ %s",
				supplier.TotalSupply.String(), calculatedTotalSupply,
			),
		)
	}

	return nil
}

// String returns a human readable string representation of a supplier.
func (supplier Supplier) String() string {
	return fmt.Sprintf(`Supplier:
  Circulating Supply:  %s
  Vesting Supply: %s
  Modules Supply:  %s
	Total Supply:  %s`,
		supplier.CirculatingSupply.String(),
		supplier.VestingSupply.String(),
		supplier.ModulesSupply.String(),
		supplier.TotalSupply.String())
}
