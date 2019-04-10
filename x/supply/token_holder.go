package supply

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TokenHolder defines an instance of a module that holds tokens
type TokenHolder struct {
	module        string
	tokenHoldings sdk.Coins
}

// NewTokenHolder creates a new TokenHolder instance
func NewTokenHolder(moduleName string, initialHoldings sdk.Coins) TokenHolder {
	return TokenHolder{
		module:        moduleName,
		tokenHoldings: initialHoldings,
	}
}

// GetHoldings returns the total tokens held by a module
func (tk TokenHolder) GetHoldings() sdk.Coins {
	return tk.tokenHoldings
}

// GetHoldingsOf returns the a total coin denom holdings retained by a module
func (tk TokenHolder) GetHoldingsOf(denom string) sdk.Int {
	return tk.tokenHoldings.AmountOf(denom)
}

// RequestTokens re
func (tk TokenHolder) RequestTokens(amount sdk.Coins) (err sdk.Error) {
	// get available supply (not held by a module or account)
	// check if supply > amount
	//
	return
}

// RelinquishTokens hands over a portion of the module's holdings
func (tk TokenHolder) RelinquishTokens(amount sdk.Coins) (err sdk.Error) {
	holdings := tk.GetHoldings()
	if !holdings.IsAllGTE(amount) {
		return ErrInsufficientHoldings(DefaultCodespace)
	}
	tk.setTokenHoldings(holdings.Sub(amount))
	return
}

// set new token holdings
func (tk *TokenHolder) setTokenHoldings(amount sdk.Coins) {
	tk.tokenHoldings = amount
}
