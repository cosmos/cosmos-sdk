package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SupplyI defines an inflationary supply interface for modules that handle
// token supply.
// Deprecated.
type SupplyI interface {
	proto.Message

	GetTotal() sdk.Coins
	SetTotal(total sdk.Coins)

	Inflate(amount sdk.Coins)
	Deflate(amount sdk.Coins)

	String() string
	ValidateBasic() error
}

// Implements Supply interface
// Deprecated.
var _ SupplyI = (*Supply)(nil)

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

// ValidateBasic validates the Supply coins and returns error if invalid
func (supply Supply) ValidateBasic() error {
	if err := supply.Total.Validate(); err != nil {
		return fmt.Errorf("invalid total supply: %w", err)
	}

	return nil
}
