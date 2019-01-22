package mock

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validator implements sdk.Validator
type Validator struct {
	Address sdk.ValAddress
	Power   sdk.Int
}

// Implements sdk.Validator
func (v Validator) GetStatus() sdk.BondStatus {
	return sdk.Bonded
}

// Implements sdk.Validator
func (v Validator) GetOperator() sdk.ValAddress {
	return v.Address
}

// Implements sdk.Validator
func (v Validator) GetConsPubKey() crypto.PubKey {
	return nil
}

// Implements sdk.Validator
func (v Validator) GetConsAddr() sdk.ConsAddress {
	return nil
}

// Implements sdk.Validator
func (v Validator) GetTokens() sdk.Int {
	return sdk.ZeroInt()
}

// Implements sdk.Validator
func (v Validator) GetPower() sdk.Int {
	return v.Power
}

// Implements sdk.Validator
func (v Validator) GetDelegatorShares() sdk.Dec {
	return sdk.ZeroDec()
}

// Implements sdk.Validator
func (v Validator) GetCommission() sdk.Dec {
	return sdk.ZeroDec()
}

// Implements sdk.Validator
func (v Validator) GetJailed() bool {
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
func (v Validator) GetDelegatorShareExRate() sdk.Dec {
	return sdk.ZeroDec()
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

// IterateBondedValidatorsByPower implements sdk.ValidatorSet
func (vs *ValidatorSet) IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, Validator sdk.Validator) bool) {
	vs.IterateValidators(ctx, fn)
}

// IterateLastValidators implements sdk.ValidatorSet
func (vs *ValidatorSet) IterateLastValidators(ctx sdk.Context, fn func(index int64, Validator sdk.Validator) bool) {
	vs.IterateValidators(ctx, fn)
}

// Validator implements sdk.ValidatorSet
func (vs *ValidatorSet) Validator(ctx sdk.Context, addr sdk.ValAddress) sdk.Validator {
	for _, val := range vs.Validators {
		if bytes.Equal(val.Address.Bytes(), addr.Bytes()) {
			return val
		}
	}
	return nil
}

// ValidatorByPubKey implements sdk.ValidatorSet
func (vs *ValidatorSet) ValidatorByConsPubKey(_ sdk.Context, _ crypto.PubKey) sdk.Validator {
	panic("not implemented")
}

// ValidatorByPubKey implements sdk.ValidatorSet
func (vs *ValidatorSet) ValidatorByConsAddr(_ sdk.Context, _ sdk.ConsAddress) sdk.Validator {
	panic("not implemented")
}

// TotalPower implements sdk.ValidatorSet
func (vs *ValidatorSet) TotalPower(ctx sdk.Context) sdk.Int {
	res := sdk.ZeroInt()
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
func (vs *ValidatorSet) Slash(_ sdk.Context, _ sdk.ConsAddress, _ int64, _ int64, _ sdk.Dec) {
	panic("not implemented")
}

// Implements sdk.ValidatorSet
func (vs *ValidatorSet) Jail(_ sdk.Context, _ sdk.ConsAddress) {
	panic("not implemented")
}

// Implements sdk.ValidatorSet
func (vs *ValidatorSet) Unjail(_ sdk.Context, _ sdk.ConsAddress) {
	panic("not implemented")
}

// Implements sdk.ValidatorSet
func (vs *ValidatorSet) Delegation(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) sdk.Delegation {
	panic("not implemented")
}
