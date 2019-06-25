package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	yaml "gopkg.in/yaml.v2"
)

// Supply represents a struct that passively keeps track of the total supply amounts in the network
type Supply struct {
	Total sdk.Coins `json:"total_supply"` // total supply of tokens registered on the chain
}

// NewSupply creates a new Supply instance
func NewSupply(total sdk.Coins) Supply { return Supply{total} }

// DefaultSupply creates an empty Supply
func DefaultSupply() Supply { return NewSupply(sdk.NewCoins()) }

// Inflate adds coins to the total supply
func (supply *Supply) Inflate(amount sdk.Coins) {
	supply.Total = supply.Total.Add(amount)
}

// Deflate subtracts coins from the total supply
func (supply *Supply) Deflate(amount sdk.Coins) {
	supply.Total = supply.Total.Sub(amount)
}

// String returns a human readable string representation of a supplier.
func (supply Supply) String() string {
	b, err := yaml.Marshal(supply)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// ValidateBasic validates the Supply coins and returns error if invalid
func (supply Supply) ValidateBasic() error {
	if !supply.Total.IsValid() {
		return fmt.Errorf("invalid total supply: %s", supply.Total.String())
	}
	return nil
}
