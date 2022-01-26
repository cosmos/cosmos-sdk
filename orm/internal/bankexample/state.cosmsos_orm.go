package bankexample

import (
	context "context"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	ormtable "github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

type StateStoreAccessor interface {
	Open(context.Context) StateStore
}

func NewStateStoreAccessor() (StateStoreAccessor, error) {
	panic("TODO")
}

type StateStore interface {
	BalanceStore
	SupplyStore
}

type BalanceStore interface {
	HasBalance(address string, denom string) (found bool, err error)
	GetBalance(address string, denom string) (*Balance, error)
	CreateBalance(balance *Balance) error
	UpdateBalance(balance *Balance) error
	SaveBalance(balance *Balance) error
	DeleteBalance(balance *Balance) error
	ListBalance(BalanceIndexKey) (BalanceIterator, error)
}

type BalanceIterator interface {
	Next() bool
	Value() (*Balance, error)
}

type BalanceIndexKey interface {
	id() uint32
	values() []protoreflect.Value
	balanceIndexKey()
}

type BalanceAddressDenomIndexKey struct {
}

type BalanceDenomAddressIndexKey struct {
}

type SupplyStore interface {
	HasSupply(denom string) (found bool, err error)
	GetSupply(denom string) (*Supply, error)
	CreateSupply(supply *Supply) error
	UpdateSupply(supply *Supply) error
	SaveSupply(supply *Supply) error
	DeleteSupply(supply *Supply) error
	ListSupply(SupplyIndexKey) (SupplyIterator, error)
}

type SupplyIterator interface {
	Next() bool
	Value() (*Supply, error)
}

type SupplyIndexKey interface {
	id() uint32
	values() []protoreflect.Value
	supplyIndexKey()
}

type SupplyDenomIndexKey struct {
}

type stateStore struct {
	balanceTable ormtable.Table
	supplyTable  ormtable.Table
}
