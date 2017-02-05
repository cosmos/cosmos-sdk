package stake

import (
	"bytes"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

type StakeState struct {
	// bonded coins
	Collateral Collaterals

	// queue of coins in unbonding period
	Unbonding []Unbond
}

type Collateral struct {
	ValidatorPubKey []byte
	Address         []byte // basecoin account which paid this collateral
	Amount          uint64
}

type Collaterals []Collateral

func (c Collaterals) Validators() []*abci.Validator {
	validators := make([]*abci.Validator, 0, len(c))
	for _, coll := range c {
		vLen := len(validators)
		if vLen == 0 || !bytes.Equal(validators[vLen-1].PubKey, coll.ValidatorPubKey) {
			v := &abci.Validator{coll.ValidatorPubKey, coll.Amount}
			validators = append(validators, v)
		} else {
			validators[vLen-1].Power += coll.Amount
		}
	}
	return validators
}

func (c Collaterals) insert(i int, coll Collateral) Collaterals {
	out := append(c, Collateral{})
	copy(out[i+1:], out[i:])
	out[i] = coll
	return out
}

// Add adds to the collateral set, sorted by ValidatorPubKey. If
// Collateral already exists with the same ValidatorPubKey AND
// Address, the amount is added.
func (c Collaterals) Add(adding Collateral) Collaterals {
	for i, coll := range c {
		validatorCmp := bytes.Compare(coll.ValidatorPubKey, adding.ValidatorPubKey)
		if validatorCmp == 1 {
			return c.insert(i, adding)
		}
		if validatorCmp == 0 {
			addressCmp := bytes.Compare(coll.Address, adding.Address)
			if addressCmp == 1 {
				return c.insert(i, adding)
			}
			if addressCmp == 0 {
				c[i].Amount += adding.Amount
				return c
			}
		}
	}
	return append(c, adding)
}

func (c Collaterals) Get(address, validatorPubKey []byte) (*Collateral, int) {
	for i, coll := range c {
		validatorCmp := bytes.Compare(coll.ValidatorPubKey, validatorPubKey)
		if validatorCmp == 1 {
			return nil, -1
		}
		if validatorCmp == 0 {
			addressCmp := bytes.Compare(coll.Address, address)
			if addressCmp == 1 {
				return nil, -1
			}
			if addressCmp == 0 {
				return &coll, i
			}
		}
	}
	return nil, -1
}

func (c Collaterals) Remove(i int) Collaterals {
	return append(c[:i], c[i+1:]...)
}

//--------------------------------------------------------------------------------

type Unbond struct {
	ValidatorPubKey []byte
	Address         []byte // basecoin account to pay out to
	Amount          uint64
	Height          uint64 // when the unbonding started
}

//--------------------------------------------------------------------------------

type Tx interface{}

type BondTx struct {
	ValidatorPubKey []byte
}

type UnbondTx struct {
	ValidatorPubKey []byte
	Amount          uint64
}

var _ = wire.RegisterInterface(
	struct{ Tx }{},
	wire.ConcreteType{BondTx{}, 0x01},
	wire.ConcreteType{UnbondTx{}, 0x02},
)
