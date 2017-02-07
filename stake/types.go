package stake

import (
	"bytes"
	"sort"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

// State defines a on-blockchain stake plugin state
type State struct {
	// bonded coins
	Collateral Collaterals

	// queue of coins in unbonding period
	Unbonding []Unbond
}

// Collateral defines one amount of bonded coins, with the public key of
// the validator who the collateral is bonded to, and the address of the
// owner of the collateral.
type Collateral struct {
	ValidatorPubKey []byte
	Address         []byte // basecoin account which paid this collateral
	Amount          uint64
}

// Collaterals is a list of Collateral objects, sorted by validator public key
// and then by address.
type Collaterals []Collateral

// Validators generates a list of abci.Validator objects based on the
// Collateral objects in the list. Validators are sorted by voting power.
func (c Collaterals) Validators() []*abci.Validator {
	validators := make(SortedValidators, 0, len(c))
	for _, coll := range c {
		vLen := len(validators)
		if vLen == 0 || !bytes.Equal(validators[vLen-1].PubKey, coll.ValidatorPubKey) {
			v := &abci.Validator{
				PubKey: coll.ValidatorPubKey,
				Power:  coll.Amount,
			}
			validators = append(validators, v)
		} else {
			validators[vLen-1].Power += coll.Amount
		}
	}
	// sort is stable because we break voting power ties using pubkey
	sort.Sort(validators)
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
// Address, the amount is added to the existing object.
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

// Get returns the collateral object with the given address and validatorPubKey,
// or nil if it doesn't exist, along with the index (or -1 if not found).
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

// Remove takes the Collateral with index `i` out of the list
func (c Collaterals) Remove(i int) Collaterals {
	return append(c[:i], c[i+1:]...)
}

// SortedValidators defines a slice of abci.Validator objects,
// which is sortable by voting power (breaking ties with public keys)
type SortedValidators []*abci.Validator

func (vals SortedValidators) Len() int {
	return len(vals)
}

func (vals SortedValidators) Less(i, j int) bool {
	morePower := vals[i].Power > vals[j].Power
	lowerKey := bytes.Compare(vals[i].PubKey, vals[j].PubKey) == -1
	return morePower || lowerKey
}

func (vals SortedValidators) Swap(i, j int) {
	vals[i], vals[j] = vals[j], vals[i]
}

//--------------------------------------------------------------------------------

// Unbond defines an amount of collateral which is in the unbonding period
type Unbond struct {
	ValidatorPubKey []byte
	Address         []byte // basecoin account to pay out to
	Amount          uint64
	Height          uint64 // when the unbonding started
}

//--------------------------------------------------------------------------------

// Tx is the interface for stake transactions
type Tx interface{}

// BondTx bonds coins as collateral
type BondTx struct {
	ValidatorPubKey []byte
}

// UnbondTx places collateral into the unbonding period
type UnbondTx struct {
	ValidatorPubKey []byte
	Amount          uint64
}

var _ = wire.RegisterInterface(
	struct{ Tx }{},
	wire.ConcreteType{BondTx{}, 0x01},
	wire.ConcreteType{UnbondTx{}, 0x02},
)
