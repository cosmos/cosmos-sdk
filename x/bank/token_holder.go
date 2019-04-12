package bank

import (
	"fmt"

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

// TokenMinter defines the interface used for modules that are allowed to mint
// and hold tokens on their behalf
type TokenMinter interface {
	TokenHolder

	GetMintedTokens() sdk.Coins
	GetMintedTokensOf(string) sdk.Int
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

//-----------------------------------------------------------------------------
// BaseTokenMinter

var _ TokenMinter = (*BaseTokenMinter)(nil)

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

// GetMintedTokens returns the a total amount minted of specific coin
func (btm BaseTokenMinter) GetMintedTokens() sdk.Coins {
	return btm.MintedTokens
}

// GetMintedTokensOf returns the a total amount minted of specific coin
func (btm BaseTokenMinter) GetMintedTokensOf(denom string) sdk.Int {
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
