package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolShares reflects the shares of a validator in a pool.
type PoolShares struct {
	Status sdk.BondStatus `json:"status"`
	Amount sdk.Rat        `json:"amount"`
}

// Equal returns a boolean determining of two PoolShares are identical.
func (s PoolShares) Equal(s2 PoolShares) bool {
	return s.Status == s2.Status &&
		s.Amount.Equal(s2.Amount)
}

// NewUnbondedShares returns a new PoolShares with a specified unbonded amount.
func NewUnbondedShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Status: sdk.Unbonded,
		Amount: amount,
	}
}

// NewUnbondingShares returns a new PoolShares with a specified unbonding
// amount.
func NewUnbondingShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Status: sdk.Unbonding,
		Amount: amount,
	}
}

// NewBondedShares returns a new PoolSahres with a specified bonding amount.
func NewBondedShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Status: sdk.Bonded,
		Amount: amount,
	}
}

// Unbonded returns the amount of unbonded shares.
func (s PoolShares) Unbonded() sdk.Rat {
	if s.Status == sdk.Unbonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// Unbonding returns the amount of unbonding shares.
func (s PoolShares) Unbonding() sdk.Rat {
	if s.Status == sdk.Unbonding {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// Bonded returns amount of bonded shares.
func (s PoolShares) Bonded() sdk.Rat {
	if s.Status == sdk.Bonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// ToUnbonded returns the equivalent amount of pool shares if the shares were
// unbonded.
func (s PoolShares) ToUnbonded(p Pool) PoolShares {
	var amount sdk.Rat

	switch s.Status {
	case sdk.Bonded:
		// (tok/bondedshr)/(tok/unbondedshr) = unbondedshr/bondedshr
		exRate := p.BondedShareExRate().Quo(p.UnbondedShareExRate())
		// bondedshr*unbondedshr/bondedshr = unbondedshr
		amount = s.Amount.Mul(exRate)
	case sdk.Unbonding:
		// (tok/unbondingshr)/(tok/unbondedshr) = unbondedshr/unbondingshr
		exRate := p.UnbondingShareExRate().Quo(p.UnbondedShareExRate())
		// unbondingshr*unbondedshr/unbondingshr = unbondedshr
		amount = s.Amount.Mul(exRate)
	case sdk.Unbonded:
		amount = s.Amount
	}

	return NewUnbondedShares(amount)
}

// ToUnbonding returns the equivalent amount of pool shares if the shares were
// unbonding.
func (s PoolShares) ToUnbonding(p Pool) PoolShares {
	var amount sdk.Rat

	switch s.Status {
	case sdk.Bonded:
		// (tok/bondedshr)/(tok/unbondingshr) = unbondingshr/bondedshr
		exRate := p.BondedShareExRate().Quo(p.UnbondingShareExRate())
		// bondedshr*unbondingshr/bondedshr = unbondingshr
		amount = s.Amount.Mul(exRate)
	case sdk.Unbonding:
		amount = s.Amount
	case sdk.Unbonded:
		// (tok/unbondedshr)/(tok/unbondingshr) = unbondingshr/unbondedshr
		exRate := p.UnbondedShareExRate().Quo(p.UnbondingShareExRate())
		// unbondedshr*unbondingshr/unbondedshr = unbondingshr
		amount = s.Amount.Mul(exRate)
	}

	return NewUnbondingShares(amount)
}

// ToBonded the equivalent amount of pool shares if the shares were bonded.
func (s PoolShares) ToBonded(p Pool) PoolShares {
	var amount sdk.Rat

	switch s.Status {
	case sdk.Bonded:
		amount = s.Amount
	case sdk.Unbonding:
		// (tok/ubshr)/(tok/bshr) = bshr/ubshr
		exRate := p.UnbondingShareExRate().Quo(p.BondedShareExRate())
		// ubshr*bshr/ubshr = bshr
		amount = s.Amount.Mul(exRate)
	case sdk.Unbonded:
		// (tok/ubshr)/(tok/bshr) = bshr/ubshr
		exRate := p.UnbondedShareExRate().Quo(p.BondedShareExRate())
		// ubshr*bshr/ubshr = bshr
		amount = s.Amount.Mul(exRate)
	}

	return NewUnbondedShares(amount)
}

// Tokens returns the equivalent amount of tokens contained by the pool shares
// for a given pool.
func (s PoolShares) Tokens(p Pool) sdk.Rat {
	switch s.Status {
	case sdk.Bonded:
		return p.BondedShareExRate().Mul(s.Amount)
	case sdk.Unbonding:
		return p.UnbondingShareExRate().Mul(s.Amount)
	case sdk.Unbonded:
		return p.UnbondedShareExRate().Mul(s.Amount)
	default:
		panic("unknown share kind")
	}
}
