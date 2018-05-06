package types

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
)

type Validator interface {
	GetAddress() Address
	GetPubKey() crypto.PubKey
	GetPower() Rat
}

func ABCIValidator(v Validator) abci.Validator {
	return abci.Validator{
		PubKey: v.GetPubKey().Bytes(),
		Power:  v.GetPower().Evaluate(),
	}
}

type ValidatorSet interface {
	Iterate(func(int, Validator))
	Size() int
}

type ValidatorSetKeeper interface {
	ValidatorSet(Context) ValidatorSet
	GetByAddress(Context, Address) Validator
	TotalPower(Context) Rat
}
