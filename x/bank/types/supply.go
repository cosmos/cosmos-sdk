package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
)

// Implements Delegation interface
var _ exported.SupplyI = (*Supply)(nil)

// NewSupply creates a new Supply instance
func NewSupply(total sdk.Coins) *Supply {
	return &Supply{total}
}

// DefaultSupply creates an empty Supply
func DefaultSupply() *Supply {
	return NewSupply(sdk.NewCoins())
}

// SetTotal sets the total supply.
func (supply *Supply) SetTotal(total sdk.Coins) {
	supply.Total = total
}

// GetTotal returns the supply total.
func (supply Supply) GetTotal() sdk.Coins {
	return supply.Total
}

// Inflate adds coins to the total supply
func (supply *Supply) Inflate(amount sdk.Coins) {
	supply.Total = supply.Total.Add(amount...)
}

// Deflate subtracts coins from the total supply.
func (supply *Supply) Deflate(amount sdk.Coins) {
	supply.Total = supply.Total.Sub(amount)
}

// String returns a human readable string representation of a supplier.
func (supply Supply) String() string {
	bz, _ := yaml.Marshal(supply)
	return string(bz)
}

// ValidateBasic validates the Supply coins and returns error if invalid
func (supply Supply) ValidateBasic() error {
	if err := supply.Total.Validate(); err != nil {
		return fmt.Errorf("invalid total supply: %w", err)
	}

	return nil
}
