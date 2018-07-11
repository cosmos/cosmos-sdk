package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Tokens of a validator
type Tokens struct {
	Status sdk.BondStatus `json:"status"`
	Amount sdk.Rat        `json:"amount"`
}

// Equal returns a boolean determining of two Tokens are identical.
func (s Tokens) Equal(s2 Tokens) bool {
	return s.Status == s2.Status &&
		s.Amount.Equal(s2.Amount)
}

// NewUnbondedTokens returns a new Tokens with a specified unbonded amount.
func NewUnbondedTokens(amount sdk.Rat) Tokens {
	return Tokens{
		Status: sdk.Unbonded,
		Amount: amount,
	}
}

// NewUnbondingTokens returns a new Tokens with a specified unbonding
// amount.
func NewUnbondingTokens(amount sdk.Rat) Tokens {
	return Tokens{
		Status: sdk.Unbonding,
		Amount: amount,
	}
}

// NewBondedTokens returns a new PoolSahres with a specified bonding amount.
func NewBondedTokens(amount sdk.Rat) Tokens {
	return Tokens{
		Status: sdk.Bonded,
		Amount: amount,
	}
}

// Unbonded returns the amount of unbonded shares.
func (s Tokens) Unbonded() sdk.Rat {
	if s.Status == sdk.Unbonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// Unbonding returns the amount of unbonding shares.
func (s Tokens) Unbonding() sdk.Rat {
	if s.Status == sdk.Unbonding {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// Bonded returns amount of bonded shares.
func (s Tokens) Bonded() sdk.Rat {
	if s.Status == sdk.Bonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}
