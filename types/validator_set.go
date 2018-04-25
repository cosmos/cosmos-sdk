package types

import (
	abci "github.com/tendermint/abci/types"
)

type Validator = abci.Validator

type ValidatorSetKeeper interface {
	Validators(Context) []*Validator
	Size(Context) int
	IsValidator(Context, Address) bool
	GetByAddress(Context, Address) (int, *Validator)
	GetByIndex(Context, int) *Validator
	TotalPower(Context) Rat
}
