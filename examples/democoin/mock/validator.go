package mock

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// Validator implements sdk.Validator
type Validator struct {
	Address sdk.AccAddress
	Power   sdk.Rat
}

// Implements sdk.Validator
func (v Validator) GetStatus() sdk.BondStatus {
	return sdk.Bonded
}

// Implements sdk.Validator
func (v Validator) GetOwner() sdk.AccAddress {
	return v.Address
}

// Implements sdk.Validator
func (v Validator) GetPubKey() crypto.PubKey {
	return nil
}

// Implements sdk.Validator
func (v Validator) GetPower() sdk.Rat {
	return v.Power
}

// Implements sdk.Validator
func (v Validator) GetDelegatorShares() sdk.Rat {
	return sdk.ZeroRat()
}

// Implements sdk.Validator
func (v Validator) GetRevoked() bool {
	return false
}

// Implements sdk.Validator
func (v Validator) GetBondHeight() int64 {
	return 0
}

// Implements sdk.Validator
func (v Validator) GetMoniker() string {
	return ""
}

// Implements sdk.Validator
type ValidatorSet struct {
	Validators []Validator
}

// IterateValidators implements sdk.ValidatorSet
func (vs *ValidatorSet) IterateValidators(ctx sdk.Context, fn func(index int64, Validator sdk.Validator) bool) {
	for i, val := range vs.Validators {
		if fn(int64(i), val) {
			break
		}
	}
}

// IterateValidatorsBonded implements sdk.ValidatorSet
func (vs *ValidatorSet) IterateValidatorsBonded(ctx sdk.Context, fn func(index int64, Validator sdk.Validator) bool) {
	vs.IterateValidators(ctx, fn)
}

// Validator implements sdk.ValidatorSet
func (vs *ValidatorSet) Validator(ctx sdk.Context, addr sdk.AccAddress) sdk.Validator {
	for _, val := range vs.Validators {
		if bytes.Equal(val.Address, addr) {
			return val
		}
	}
	return nil
}

// TotalPower implements sdk.ValidatorSet
func (vs *ValidatorSet) TotalPower(ctx sdk.Context) sdk.Rat {
	res := sdk.ZeroRat()
	for _, val := range vs.Validators {
		res = res.Add(val.Power)
	}
	return res
}

// Helper function for adding new validator
func (vs *ValidatorSet) AddValidator(val Validator) {
	vs.Validators = append(vs.Validators, val)
}

// Helper function for removing exsting validator
func (vs *ValidatorSet) RemoveValidator(addr sdk.AccAddress) {
	pos := -1
	for i, val := range vs.Validators {
		if bytes.Equal(val.Address, addr) {
			pos = i
			break
		}
	}
	if pos == -1 {
		return
	}
	vs.Validators = append(vs.Validators[:pos], vs.Validators[pos+1:]...)
}

// Implements sdk.ValidatorSet
func (vs *ValidatorSet) Slash(ctx sdk.Context, pubkey crypto.PubKey, height int64, power int64, amt sdk.Rat) {
	panic("not implemented")
}

// Implements sdk.ValidatorSet
func (vs *ValidatorSet) Revoke(ctx sdk.Context, pubkey crypto.PubKey) {
	panic("not implemented")
}

// Implements sdk.ValidatorSet
func (vs *ValidatorSet) Unrevoke(ctx sdk.Context, pubkey crypto.PubKey) {
	panic("not implemented")
}
