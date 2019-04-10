package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TokenHolder defines an instance of a module that holds tokens
type TokenHolder struct {
	Module   string    `json:"module"`
	Holdings sdk.Coins `json:"holdings"`
}

// NewTokenHolder creates a new TokenHolder instance
func NewTokenHolder(moduleName string, initialHoldings sdk.Coins) TokenHolder {
	return TokenHolder{
		Module:   moduleName,
		Holdings: initialHoldings,
	}
}

// HoldingsOf returns the a total coin denom holdings retained by a module
func (tk TokenHolder) HoldingsOf(denom string) sdk.Int {
	return tk.Holdings.AmountOf(denom)
}

// RequestTokens adds requested tokens to the module's holdings
func (tk TokenHolder) RequestTokens(amount sdk.Coins) error {
	// get available supply (not held by a module or account)
	availableSupply := keeper.GetCirculatingSupply(ctx) // TODO: move to keeper ?
	if !availableSupply.IsAllGTE(amount) {
		return fmt.Errorf("requested tokens greater than current circulating free supply")
	}
	tk.setTokenHoldings(tk.Holdings.Add(amount))
	return nil
}

// RelinquishTokens hands over a portion of the module's holdings
func (tk TokenHolder) RelinquishTokens(amount sdk.Coins) error {
	if !tk.Holdings.IsAllGTE(amount) {
		return fmt.Errorf("insufficient token holdings")
	}
	tk.setTokenHoldings(tk.Holdings.Sub(amount))
	return nil
}

// set new token holdings
func (tk *TokenHolder) setTokenHoldings(amount sdk.Coins) {
	tk.Holdings = amount
}
