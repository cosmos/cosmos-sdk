package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// define consts for inflatoin
const (
	InflateCirculating = "circulating"
	InflateVesting     = "vesting"
	InflateHolders     = "holders"
)

// Supplier represents the keeps track of the total supply amounts in the network
type Supplier struct {
	CirculatingSupply sdk.Coins `json:"circulating_supply"` // supply held by accounts that's not vesting
	VestingSupply     sdk.Coins `json:"vesting_supply"`     // locked supply held by vesting accounts
	HoldersSupply     sdk.Coins `json:"holders_supply"`     // supply held by non acccount token holders (e.g modules)
	TotalSupply       sdk.Coins `json:"total_supply"`       // total supply of the network
}

// NewSupplier creates a new Supplier instance
func NewSupplier(circulating, vesting, holders, total sdk.Coins) Supplier {
	return Supplier{
		CirculatingSupply: circulating,
		VestingSupply:     vesting,
		HoldersSupply:     holders,
		TotalSupply:       total,
	}
}

// DefaultSupplier creates an empty Supplier
func DefaultSupplier() Supplier {
	return NewSupplier(sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, sdk.Coins{})
}

// Inflate returns the circulating supply of a coin denomination
func (supplier *Supplier) Inflate(inflationType string, amt sdk.Coins) error {
	switch inflationType {
	case InflateCirculating:
		supplier.CirculatingSupply = supplier.CirculatingSupply.Add(amt)
		break
	case InflateHolders:
		supplier.HoldersSupply = supplier.HoldersSupply.Add(amt)
		break
	case InflateVesting:
		supplier.VestingSupply = supplier.VestingSupply.Add(amt)
		break
	default:
		return fmt.Errorf("invalid type %s", inflationType)
	}
	supplier.TotalSupply = supplier.TotalSupply.Add(amt)
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

// HoldersAmountOf returns the total token holders' supply of a coin denomination
func (supplier Supplier) HoldersAmountOf(denom string) sdk.Int {
	return supplier.HoldersSupply.AmountOf(denom)
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
	if !supplier.HoldersSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid token holders supply: %s", supplier.HoldersSupply.String()),
		)
	}
	if !supplier.TotalSupply.IsValid() {
		return sdk.ErrInvalidCoins(
			fmt.Sprintf("invalid total supply: %s", supplier.TotalSupply.String()),
		)
	}

	calculatedTotalSupply :=
		supplier.CirculatingSupply.Add(supplier.VestingSupply).Add(supplier.HoldersSupply)

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
  Holders Supply:  %s
	Total Supply:  %s`,
		supplier.CirculatingSupply.String(),
		supplier.VestingSupply.String(),
		supplier.HoldersSupply.String(),
		supplier.TotalSupply.String())
}
