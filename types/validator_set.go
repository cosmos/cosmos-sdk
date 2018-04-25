package types

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/wire"
)

var cdc = wire.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
}

type Validator = abci.Validator

type ValidatorSetKeeper interface {
	Hash(Context) []byte
	GetValidators(Context) []*Validator
	Size(Context) int
	IsValidator(Context, Address) bool
	GetByAddress(Context, Address) (int, *Validator)
	GetByIndex(Context, int) *Validator
	TotalPower(Context) Rat
}
