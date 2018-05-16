package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// kind of shares
type PoolShareKind byte

// nolint
const (
	ShareUnbonded  PoolShareKind = 0x00
	ShareUnbonding PoolShareKind = 0x01
	ShareBonded    PoolShareKind = 0x02
)

// pool shares held by a validator
type PoolShares struct {
	Kind   PoolShareKind `json:"kind"`
	Amount sdk.Rat       `json:"shares"` // total shares of type ShareKind
}

// only the vitals - does not check bond height of IntraTxCounter
func (s PoolShares) Equal(s2 PoolShares) bool {
	return s.Kind == s2.Kind &&
		s.Amount.Equal(s2.Amount)
}

func NewUnbondedShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Kind:   ShareUnbonded,
		Amount: amount,
	}
}

func NewUnbondingShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Kind:   ShareUnbonding,
		Amount: amount,
	}
}

func NewBondedShares(amount sdk.Rat) PoolShares {
	return PoolShares{
		Kind:   ShareBonded,
		Amount: amount,
	}
}

//_________________________________________________________________________________________________________

// amount of unbonded shares
func (s PoolShares) Unbonded() sdk.Rat {
	if s.Kind == ShareUnbonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// amount of unbonding shares
func (s PoolShares) Unbonding() sdk.Rat {
	if s.Kind == ShareUnbonding {
		return s.Amount
	}
	return sdk.ZeroRat()
}

// amount of bonded shares
func (s PoolShares) Bonded() sdk.Rat {
	if s.Kind == ShareBonded {
		return s.Amount
	}
	return sdk.ZeroRat()
}

//_________________________________________________________________________________________________________

// equivalent amount of shares if the shares were unbonded
func (s PoolShares) ToUnbonded(p Pool) PoolShares {
	var amount sdk.Rat
	switch s.Kind {
	case ShareBonded:
		exRate := p.bondedShareExRate().Quo(p.unbondedShareExRate()) // (tok/bondedshr)/(tok/unbondedshr) = unbondedshr/bondedshr
		amount = s.Amount.Mul(exRate)                                // bondedshr*unbondedshr/bondedshr = unbondedshr
	case ShareUnbonding:
		exRate := p.unbondingShareExRate().Quo(p.unbondedShareExRate()) // (tok/unbondingshr)/(tok/unbondedshr) = unbondedshr/unbondingshr
		amount = s.Amount.Mul(exRate)                                   // unbondingshr*unbondedshr/unbondingshr = unbondedshr
	case ShareUnbonded:
		amount = s.Amount
	}
	return NewUnbondedShares(amount)
}

// equivalent amount of shares if the shares were unbonding
func (s PoolShares) ToUnbonding(p Pool) PoolShares {
	var amount sdk.Rat
	switch s.Kind {
	case ShareBonded:
		exRate := p.bondedShareExRate().Quo(p.unbondingShareExRate()) // (tok/bondedshr)/(tok/unbondingshr) = unbondingshr/bondedshr
		amount = s.Amount.Mul(exRate)                                 // bondedshr*unbondingshr/bondedshr = unbondingshr
	case ShareUnbonding:
		amount = s.Amount
	case ShareUnbonded:
		exRate := p.unbondedShareExRate().Quo(p.unbondingShareExRate()) // (tok/unbondedshr)/(tok/unbondingshr) = unbondingshr/unbondedshr
		amount = s.Amount.Mul(exRate)                                   // unbondedshr*unbondingshr/unbondedshr = unbondingshr
	}
	return NewUnbondingShares(amount)
}

// equivalent amount of shares if the shares were bonded
func (s PoolShares) ToBonded(p Pool) PoolShares {
	var amount sdk.Rat
	switch s.Kind {
	case ShareBonded:
		amount = s.Amount
	case ShareUnbonding:
		exRate := p.unbondingShareExRate().Quo(p.bondedShareExRate()) // (tok/ubshr)/(tok/bshr) = bshr/ubshr
		amount = s.Amount.Mul(exRate)                                 // ubshr*bshr/ubshr = bshr
	case ShareUnbonded:
		exRate := p.unbondedShareExRate().Quo(p.bondedShareExRate()) // (tok/ubshr)/(tok/bshr) = bshr/ubshr
		amount = s.Amount.Mul(exRate)                                // ubshr*bshr/ubshr = bshr
	}
	return NewUnbondedShares(amount)
}

//_________________________________________________________________________________________________________

// get the equivalent amount of tokens contained by the shares
func (s PoolShares) Tokens(p Pool) sdk.Rat {
	switch s.Kind {
	case ShareBonded:
		return p.unbondedShareExRate().Mul(s.Amount) // (tokens/shares) * shares
	case ShareUnbonding:
		return p.unbondedShareExRate().Mul(s.Amount)
	case ShareUnbonded:
		return p.unbondedShareExRate().Mul(s.Amount)
	}
	return sdk.ZeroRat()
}
