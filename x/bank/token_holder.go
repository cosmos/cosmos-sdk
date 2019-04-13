package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TokenHolder defines the interface used for modules that are allowed to hold
// tokens. This is designed to prevent held tokens to be kept in regular Accounts,
// as those are only ment for user accounts
//
// The bank module keeps track of each of the module's holdings and uses it to
// calculate the total supply
type TokenHolder interface {
	GetModuleName() string

	GetHoldings() sdk.Coins
	SetHoldings(sdk.Coins)

	GetHoldingsOf(string) sdk.Int
}

//-----------------------------------------------------------------------------
// BaseTokenHolder

var _ TokenHolder = (*BaseTokenHolder)(nil)

// BaseTokenHolder defines an instance of a module that holds tokens
type BaseTokenHolder struct {
	Module   string    `json:"module"`
	Holdings sdk.Coins `json:"holdings"` // holdings from free available supply (not held by modules or accounts)
}

// NewBaseTokenHolder creates a new BaseTokenHolder instance
func NewBaseTokenHolder(moduleName string, initialHoldings sdk.Coins) BaseTokenHolder {
	return BaseTokenHolder{
		Module:   moduleName,
		Holdings: initialHoldings,
	}
}

// GetModuleName returns the the name of the holder's module
func (bth BaseTokenHolder) GetModuleName() string {
	return bth.Module
}

// GetHoldings returns the a total coin denom holdings retained by a module
func (bth BaseTokenHolder) GetHoldings() sdk.Coins {
	return bth.Holdings
}

// GetHoldings returns the a total coin denom holdings retained by a module
func (bth *BaseTokenHolder) SetHoldings(amount sdk.Coins) {
	bth.Holdings = amount
}

// GetHoldingsOf returns the a total coin denom holdings retained by a module
func (bth BaseTokenHolder) GetHoldingsOf(denom string) sdk.Int {
	return bth.Holdings.AmountOf(denom)
}
