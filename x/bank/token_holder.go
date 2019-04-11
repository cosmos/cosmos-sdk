package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TokenHolder defines the BaseTokenHolder interface
type TokenHolder interface {
	HoldingsOf()
	RequestTokens()
	RelinquishTokens()
}

// TokenMinter defines the BaseTokenMinter interface
type TokenMinter interface {
	TokenHolder

	MintTokens()
	BurnTokens()
}

//-----------------------------------------------------------------------------
// BaseTokenHolder

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

// HoldingsOf returns the a total coin denom holdings retained by a module
func (btk BaseTokenHolder) HoldingsOf(denom string) sdk.Int {
	return btk.Holdings.AmountOf(denom)
}

// RequestTokens adds requested tokens to the module's holdings
func (btk BaseTokenHolder) RequestTokens(amount sdk.Coins) error {
	if !amount.IsValid() {
		return fmt.Errorf("invalid requested amount")
	}
	// get available supply (not held by a module or account)
	// availableSupply := keeper.GetCirculatingSupply(ctx) // TODO: move to keeper ?
	// if !availableSupply.IsAllGTE(amount) {
	// 	return fmt.Errorf("requested tokens greater than current circulating free supply")
	// }
	// TODO: decrease free supply
	btk.setTokenHoldings(btk.Holdings.Add(amount))
	return nil
}

// RelinquishTokens hands over a portion of the module's holdings
func (btk BaseTokenHolder) RelinquishTokens(amount sdk.Coins) error {
	if !amount.IsValid() {
		return fmt.Errorf("invalid provided relenquished amount")
	}
	if !btk.Holdings.IsAllGTE(amount) {
		return fmt.Errorf("insufficient token holdings")
	}
	btk.setTokenHoldings(btk.Holdings.Sub(amount))
	//TODO: add to free supply
	return nil
}

// set new token holdings
func (btk *BaseTokenHolder) setTokenHoldings(amount sdk.Coins) {
	btk.Holdings = amount
}

//-----------------------------------------------------------------------------
// BaseTokenMinter

// BaseTokenMinter defines an instance of a module that is allowed to hold and mint tokens
type BaseTokenMinter struct {
	*BaseTokenHolder

	MintedTokens sdk.Coins `json:"minted_tokens"` // TODO: should this have an allowance ?
}

// NewBaseTokenMinter creates a new BaseTokenMinter instance
func NewBaseTokenMinter(moduleName string, initialHoldings, initialMintedTokens sdk.Coins,
) BaseTokenMinter {

	baseTokenHoler := NewBaseTokenHolder(moduleName, initialHoldings)
	return BaseTokenMinter{
		BaseTokenHolder: &baseTokenHoler,
		MintedTokens:    initialMintedTokens,
	}
}

// MintedTokensOf returns the a total amount minted of specific coin
func (btm BaseTokenMinter) MintedTokensOf(denom string) sdk.Int {
	return btm.MintedTokens.AmountOf(denom)
}

// Mint creates new tokens and registers them to the module
func (btm BaseTokenMinter) Mint(amount sdk.Coins) error {
	if !amount.IsValid() {
		return fmt.Errorf("invalid provided minting amount")
	}
	// TODO: Check for allowance ?
	btm.setMintedTokens(btm.MintedTokens.Add(amount))
	return nil
}

// BurnTokens destroys a portion of the previously minted tokens
func (btm BaseTokenMinter) BurnTokens(amount sdk.Coins) error {
	if !amount.IsValid() {
		return fmt.Errorf("invalid provided burning amount")
	}
	if !btm.MintedTokens.IsAllGTE(amount) {
		return fmt.Errorf("can't burn more tokens than current minted token balance")
	}

	btm.setMintedTokens(btm.MintedTokens.Sub(amount))
	return nil
}

// set new minted tokens
func (btm *BaseTokenMinter) setMintedTokens(amount sdk.Coins) {
	btm.MintedTokens = amount
}
