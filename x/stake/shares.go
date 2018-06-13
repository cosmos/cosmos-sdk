package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// kind of shares
type PoolShareKind byte

// pool shares held by a validator
type PoolShares struct {
	Status sdk.BondStatus `json:"status"`
	Amount sdk.Rat        `json:"amount"` // total shares of type ShareKind
}

// only the vitals - does not check bond height of IntraTxCounter
func (s PoolShares) Equal(s2 PoolShares) bool {
	return s.Status == s2.Status &&
		s.Amount.Equal(s2.Amount)
}

func NewUnbondedShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Status: sdk.Unbonded,
		Amount: amount,
	}
}

func NewUnbondingShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Status: sdk.Unbonding,
		Amount: amount,
	}
}

func NewBondedShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Status: sdk.Bonded,
		Amount: amount,
	}
}

//_________________________________________________________________________________________________________

// amount of unbonded shares
func (s PoolShares) Unbonded() sdk.Rat {
	if s.Status == sdk.Unbonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// amount of unbonding shares
func (s PoolShares) Unbonding() sdk.Rat {
	if s.Status == sdk.Unbonding {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// amount of bonded shares
func (s PoolShares) Bonded() sdk.Rat {
	if s.Status == sdk.Bonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}

//_________________________________________________________________________________________________________

// equivalent amount of shares if the shares were unbonded
func (s PoolShares) ToUnbonded(p Pool) PoolShares {
	var amount sdk.Rat
	switch s.Status {
	case sdk.Bonded:
		exRate := p.bondedShareExRate().Quo(p.unbondedShareExRate()) // (tok/bondedshr)/(tok/unbondedshr) = unbondedshr/bondedshr
		amount = s.Amount.Mul(exRate)                                // bondedshr*unbondedshr/bondedshr = unbondedshr
	case sdk.Unbonding:
		exRate := p.unbondingShareExRate().Quo(p.unbondedShareExRate()) // (tok/unbondingshr)/(tok/unbondedshr) = unbondedshr/unbondingshr
		amount = s.Amount.Mul(exRate)                                   // unbondingshr*unbondedshr/unbondingshr = unbondedshr
	case sdk.Unbonded:
		amount = s.Amount
	}
	return NewUnbondedShares(amount)
}

// equivalent amount of shares if the shares were unbonding
func (s PoolShares) ToUnbonding(p Pool) PoolShares {
	var amount sdk.Rat
	switch s.Status {
	case sdk.Bonded:
		exRate := p.bondedShareExRate().Quo(p.unbondingShareExRate()) // (tok/bondedshr)/(tok/unbondingshr) = unbondingshr/bondedshr
		amount = s.Amount.Mul(exRate)                                 // bondedshr*unbondingshr/bondedshr = unbondingshr
	case sdk.Unbonding:
		amount = s.Amount
	case sdk.Unbonded:
		exRate := p.unbondedShareExRate().Quo(p.unbondingShareExRate()) // (tok/unbondedshr)/(tok/unbondingshr) = unbondingshr/unbondedshr
		amount = s.Amount.Mul(exRate)                                   // unbondedshr*unbondingshr/unbondedshr = unbondingshr
	}
	return NewUnbondingShares(amount)
}

// equivalent amount of shares if the shares were bonded
func (s PoolShares) ToBonded(p Pool) PoolShares {
	var amount sdk.Rat
	switch s.Status {
	case sdk.Bonded:
		amount = s.Amount
	case sdk.Unbonding:
		exRate := p.unbondingShareExRate().Quo(p.bondedShareExRate()) // (tok/ubshr)/(tok/bshr) = bshr/ubshr
		amount = s.Amount.Mul(exRate)                                 // ubshr*bshr/ubshr = bshr
	case sdk.Unbonded:
		exRate := p.unbondedShareExRate().Quo(p.bondedShareExRate()) // (tok/ubshr)/(tok/bshr) = bshr/ubshr
		amount = s.Amount.Mul(exRate)                                // ubshr*bshr/ubshr = bshr
	}
	return NewUnbondedShares(amount)
}

//_________________________________________________________________________________________________________

// TODO better tests
// get the equivalent amount of tokens contained by the shares
func (s PoolShares) Tokens(p Pool) sdk.Rat {
	switch s.Status {
	case sdk.Bonded:
		return p.bondedShareExRate().Mul(s.Amount) // (tokens/shares) * shares
	case sdk.Unbonding:
		return p.unbondingShareExRate().Mul(s.Amount)
	case sdk.Unbonded:
		return p.unbondedShareExRate().Mul(s.Amount)
	default:
		panic("unknown share kind")
	}
}
